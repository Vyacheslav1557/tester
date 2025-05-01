package sessions

import (
	"context"
	"github.com/Vyacheslav1557/tester/internal/models"
)

type UseCase interface {
	CreateSession(ctx context.Context, creation *models.Session) error
	ReadSession(ctx context.Context, sessionId string) (*models.Session, error)
	UpdateSession(ctx context.Context, sessionId string) error
	DeleteSession(ctx context.Context, sessionId string) error
	DeleteAllSessions(ctx context.Context, userId int32) error
}
