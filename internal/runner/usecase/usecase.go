package usecase

import (
	"context"
	"github.com/Vyacheslav1557/tester/internal/runner/usecase/tester"
	"github.com/docker/docker/client"
)

type UseCase struct {
	t *tester.Tester
}

func NewUseCase(dockerClient *client.Client, cacheDir string) *UseCase {
	t := tester.NewTester(cacheDir, tester.NewDockerExecutor(dockerClient), 2)

	return &UseCase{t: t}
}

func (u *UseCase) Test(ctx context.Context, packet tester.Packet) <-chan tester.TestingMessage {
	return u.t.Test(ctx, packet)
}
