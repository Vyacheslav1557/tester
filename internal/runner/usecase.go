package runner

import (
	"context"
	"github.com/Vyacheslav1557/tester/internal/runner/usecase/tester"
)

type UseCase interface {
	Test(ctx context.Context, packet tester.Packet) <-chan tester.TestingMessage
}
