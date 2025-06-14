package tester

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/Vyacheslav1557/tester/pkg"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Executor interface {
	Compile(ctx context.Context, cfg Config, path string) error
	Execute(ctx context.Context, cfg Config, path string, input io.Reader) error
}

type DockerExecutor struct {
	dockerClient *client.Client
}

func NewDockerExecutor(dockerClient *client.Client) *DockerExecutor {
	return &DockerExecutor{dockerClient: dockerClient}
}

func (e *DockerExecutor) Compile(ctx context.Context, cfg Config, workDir string) error {
	const op = "DockerExecutor.Compile"

	var pidsLimit int64 = 100
	resp, err := e.dockerClient.ContainerCreate(ctx,
		&container.Config{
			Image:           cfg.Image(),
			Cmd:             cfg.CompileCMD(),
			Tty:             false,
			OpenStdin:       false,
			NetworkDisabled: true,
			User:            "1000:1000",
		},
		&container.HostConfig{
			Binds: []string{
				fmt.Sprintf("%s:/code:rw", workDir),
			},
			Resources: container.Resources{
				Memory:    cfg.CompileML(),
				CPUPeriod: 100000,
				CPUQuota:  100000,
				PidsLimit: &pidsLimit,
			},
			SecurityOpt: []string{"apparmor:docker-default"},
		},
		nil,
		nil,
		"",
	)
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "failed to create container")
	}

	defer e.dockerClient.ContainerRemove(ctx, resp.ID, container.RemoveOptions{})

	err = e.dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "failed to start container")
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, cfg.CompileTL())
	defer cancel()

	statusCh, errCh := e.dockerClient.ContainerWait(timeoutCtx, resp.ID, container.WaitConditionNotRunning)
	var statusCode int64

	select {
	case err := <-errCh:
		if err != nil {
			err = errors.Join(e.dockerClient.ContainerKill(ctx, resp.ID, "SIGKILL"), err)
			return pkg.Wrap(pkg.ErrInternal, err, op, "failed to kill container")
		}
	case status := <-statusCh:
		statusCode = status.StatusCode
	}

	logs, err := e.dockerClient.ContainerLogs(ctx, resp.ID, container.LogsOptions{
		ShowStderr: true,
		ShowStdout: false,
	})
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "failed to capture logs")
	}
	defer logs.Close()

	var stderrBuf strings.Builder
	_, err = stdcopy.StdCopy(nil, &stderrBuf, logs)
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "failed to read logs")
	}
	stderrOutput := stderrBuf.String()

	if statusCode != 0 || len(stderrOutput) > 0 {
		return pkg.Wrap(CompilationErr, nil, op, "compile error: "+stderrOutput)
	}

	return nil
}

func (e *DockerExecutor) Execute(ctx context.Context, cfg Config, workDir string, in io.Reader) error {
	const op = "DockerExecutor.Execute"

	var pidsLimit int64 = 100
	resp, err := e.dockerClient.ContainerCreate(ctx,
		&container.Config{
			Image:           cfg.Image(),
			Cmd:             cfg.ExecuteCMD(),
			Tty:             false,
			OpenStdin:       true,
			StdinOnce:       true,
			NetworkDisabled: true,
			User:            "1000:1000",
		},
		&container.HostConfig{
			Binds: []string{
				fmt.Sprintf("%s:/code:ro", workDir),
			},
			CapDrop: []string{
				"ALL",
			},
			CapAdd: []string{
				"SYS_CHROOT",
			},
			Resources: container.Resources{
				Memory:    256 * 1024 * 1024,
				CPUPeriod: 100000,
				CPUQuota:  100000,
				PidsLimit: &pidsLimit,
			},
			SecurityOpt: []string{
				"apparmor:docker-default",
			},
			ReadonlyRootfs: true,
		},
		nil,
		nil,
		"",
	)
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "failed to create container")
	}

	containerID := resp.ID
	defer e.dockerClient.ContainerRemove(ctx, containerID, container.RemoveOptions{})

	conn, err := e.dockerClient.ContainerAttach(ctx, containerID, container.AttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "failed to attach to container")
	}
	defer conn.Close()

	err = e.dockerClient.ContainerStart(ctx, containerID, container.StartOptions{})
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "failed to start container")
	}

	_, err = io.Copy(conn.Conn, in)
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "failed to write to container")
	}

	err = conn.CloseWrite()
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "failed to close write to container")
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	outputDone := make(chan error)
	go func() {
		_, err := stdcopy.StdCopy(&stdoutBuf, &stderrBuf, conn.Reader)
		outputDone <- err
	}()

	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second) // FIXME: make this configurable
	defer cancel()

	statusCh, errCh := e.dockerClient.ContainerWait(timeoutCtx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			err = errors.Join(e.dockerClient.ContainerKill(ctx, containerID, "SIGKILL"), err)
			return pkg.Wrap(pkg.ErrInternal, err, op, "failed to wait container")
		}
	case status := <-statusCh:
		if status.StatusCode != 0 {
			err = fmt.Errorf("non-zero exit status: %d, failed to run code: %s",
				status.StatusCode, stderrBuf.String())
			return pkg.Wrap(RuntimeErr, err, op, "failed to run code")
		}
	}

	err = <-outputDone
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "failed to read logs")
	}

	timeFilePath := filepath.Join(workDir, "time.txt")
	timeFile, err := os.Create(timeFilePath)
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "failed to create time file")
	}
	defer timeFile.Close()

	_, err = io.Copy(timeFile, &stderrBuf)
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "failed to write time file")
	}

	outputFilePath := filepath.Join(workDir, "output.txt")
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, &stdoutBuf)
	if err != nil {
		return fmt.Errorf("failed to write output file: %v", err)
	}

	return nil
}
