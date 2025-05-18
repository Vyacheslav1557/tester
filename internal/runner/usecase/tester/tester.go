package tester

import (
	"context"
	"errors"
	"fmt"
	models "github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/pkg"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

type Tester struct {
	pool     *Pool[ExecuteMessage]
	cacheDir string
	compiler Compiler
}

func NewTester(cacheDir string, executor Executor, n int) *Tester {
	t := &Tester{
		cacheDir: cacheDir,
		compiler: executor,
	}

	t.pool = NewPool[ExecuteMessage](n, t.newExecutorWrapper(executor))

	return t
}

type Compiler interface {
	Compile(ctx context.Context, cfg Config, path string) error
}

type Packet interface {
	Solution() []byte
	UniquePacketName() string
	Lang() models.LanguageName
	ZipPath() string
	TL() int64
	ML() int64
	Meta() *models.Meta
}

type TestingMessage struct {
	Metrics *Metrics
	Err     error
	Details string
}

type ExecuteMessage struct {
	callback func(msg TestingMessage)
	ctx      context.Context
	cfg      Config
	workDir  string
	in       io.Reader
}

func (t *Tester) newExecutorWrapper(executor Executor) func(ExecuteMessage) {
	return func(msg ExecuteMessage) {
		err := executor.Execute(msg.ctx, msg.cfg, msg.workDir, msg.in)
		msg.callback(TestingMessage{Err: err})
	}
}

func (t *Tester) prepareTests(p Packet) (string, error) {
	zipFile, err := os.Open(p.ZipPath())
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	testsPath := filepath.Join(t.cacheDir, "tests", p.UniquePacketName())
	exists, err := pathExists(testsPath)
	if err != nil {
		return "", err
	}

	if !exists {
		if err := os.MkdirAll(testsPath, 0755); err != nil {
			return "", err
		}

		stat, err := zipFile.Stat()
		if err != nil {
			return "", err
		}

		err = unzipArchive(zipFile, stat.Size(), testsPath)
		if err != nil {
			return "", err
		}
	}

	return testsPath, nil
}

func (t *Tester) prepareSource(p Packet, workDir string) (string, error) {
	sourcePath := filepath.Join(workDir, "source")

	file, err := os.OpenFile(sourcePath, os.O_CREATE, 0600)
	if err != nil {
		return "", err
	}

	_, err = file.Write(p.Solution())
	if err != nil {
		return "", err
	}

	return sourcePath, nil
}

func (t *Tester) prepareBuild(testDir, buildPath string) (string, error) {
	const op = "Tester.prepareBuild"

	build, err := os.OpenFile(buildPath, os.O_RDONLY, 0600)
	if err != nil {
		return "", pkg.Wrap(pkg.ErrInternal, err, op, "failed to open build")
	}
	defer build.Close()

	buildCopyPath := filepath.Join(testDir, "solution")

	buildCopy, err := os.Create(buildCopyPath)
	if err != nil {
		return "", pkg.Wrap(pkg.ErrInternal, err, op, "failed to create build copy")
	}
	defer buildCopy.Close()

	_, err = io.Copy(buildCopy, build)
	if err != nil {
		return "", pkg.Wrap(pkg.ErrInternal, err, op, "failed to copy build")
	}

	return buildCopyPath, nil
}

func (t *Tester) test(p Packet, buildPath, testsPath, testName string) (*Metrics, error) {
	const op = "Tester.test"

	testDir, err := os.MkdirTemp("", "test")
	if err != nil {
		return nil, pkg.Wrap(pkg.ErrInternal, err, op, "failed to create test dir")
	}

	_, err = t.prepareBuild(testDir, buildPath)
	if err != nil {
		return nil, err
	}

	tests, err := os.OpenFile(testsPath, os.O_RDONLY, 0600)
	if err != nil {
		return nil, pkg.Wrap(pkg.ErrInternal, err, op, "failed to open tests")
	}
	defer tests.Close()

	in, err := os.OpenFile(filepath.Join(testsPath, "tests", testName), os.O_RDONLY, 0600)
	if err != nil {
		return nil, pkg.Wrap(pkg.ErrInternal, err, op, "failed to open test")
	}
	defer in.Close()

	ch := make(chan TestingMessage)
	defer close(ch)

	ctx := context.TODO()

	err = t.pool.Do(ctx, ExecuteMessage{
		callback: func(msg TestingMessage) {
			ch <- msg
		},
		ctx:     ctx,
		cfg:     GetConfig(p.Lang()),
		workDir: testDir,
		in:      in,
	})

	if err != nil {
		return nil, pkg.Wrap(pkg.ErrInternal, err, op, "failed to execute test")
	}

	select {
	case msg := <-ch:
		if msg.Err != nil {
			return nil, msg.Err
		}
		break
	case <-ctx.Done():
		return nil, pkg.Wrap(pkg.ErrInternal, nil, op, "timeout")
	}

	metricsPath := filepath.Join(testDir, "time.txt")
	metrics, err := parseMetrics(metricsPath)
	if err != nil {
		return nil, pkg.Wrap(pkg.ErrInternal, err, op, "failed to parse metrics")
	}

	if metrics.ElapsedTime.Milliseconds() > p.TL() {
		return metrics, pkg.Wrap(TimeLimitExceededErr, nil, op, "time limit exceeded")
	}

	if int64(metrics.MaximumResidentSetSize) >= p.ML()*1024 {
		return metrics, pkg.Wrap(MemoryLimitExceededErr, nil, op, "memory limit exceeded")
	}

	expected := filepath.Join(testsPath, "tests", testName+".a")
	actual := filepath.Join(testDir, "output.txt")

	err = compareFiles(expected, actual, 1e-6)
	if err != nil {
		return metrics, pkg.Wrap(nil, err, op, "failed to compare files")
	}

	return metrics, nil
}

func (t *Tester) Test(ctx context.Context, packet Packet) <-chan TestingMessage {
	const op = "Tester.Test"

	ch := make(chan TestingMessage)
	go func() {
		defer close(ch)

		ch <- TestingMessage{Details: "Preparing"}

		testsPath, err := t.prepareTests(packet)
		if err != nil {
			ch <- TestingMessage{
				Err: pkg.Wrap(pkg.ErrInternal, err, op, "failed to prepare tests"),
			}
			return
		}

		lang := GetConfig(packet.Lang())
		if lang == nil {
			ch <- TestingMessage{
				Err: pkg.Wrap(pkg.ErrInternal, nil, op, "unknown language"),
			}
			return
		}

		workDir, err := os.MkdirTemp("", "tester")
		if err != nil {
			ch <- TestingMessage{
				Err: pkg.Wrap(pkg.ErrInternal, err, op, "failed to create work dir"),
			}
			return
		}
		defer os.RemoveAll(workDir)

		_, err = t.prepareSource(packet, workDir)
		if err != nil {
			ch <- TestingMessage{
				Err: pkg.Wrap(pkg.ErrInternal, err, op, "failed to prepare source"),
			}
			return
		}

		compileCmd := lang.CompileCMD()
		if len(compileCmd) > 0 {
			ch <- TestingMessage{Details: "Compiling"}

			err = t.compiler.Compile(ctx, lang, workDir)
			if err != nil {
				ch <- TestingMessage{
					Err: pkg.Wrap(nil, err, op, "failed to compile"),
				}
				return
			}
		}

		buildPath := filepath.Join(workDir, "solution")

		ch <- TestingMessage{Details: "Testing"}

		meta := packet.Meta()

		wg := sync.WaitGroup{}
		wg.Add(meta.Count)

		for j := 0; j < meta.Count; j++ {
			testName := meta.Names[j]
			ch <- TestingMessage{Details: fmt.Sprintf("Testing %s", testName)}

			go func() {
				defer wg.Done()

				metrics, err := t.test(packet, buildPath, testsPath, testName)
				if err != nil {
					ch <- TestingMessage{
						Metrics: metrics,
						Err:     pkg.Wrap(pkg.ErrInternal, err, op, "failed to test"),
					}
					return
				}

				ch <- TestingMessage{
					Metrics: metrics,
					Details: fmt.Sprintf("%s passed", testName),
				}
			}()
		}

		wg.Wait()
	}()

	return ch
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}
