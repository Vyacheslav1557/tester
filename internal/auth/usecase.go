package auth

import (
	"context"
	"github.com/Vyacheslav1557/tester/internal/models"
)

type UseCase interface {
	Login(ctx context.Context, credentials *models.Credentials, device *models.Device) (*models.Session, error)
	Refresh(ctx context.Context, sessionId string) error
	Logout(ctx context.Context, sessionId string) error
	Terminate(ctx context.Context, userId int32) error
	ListSessions(ctx context.Context, userId int32) ([]*models.Session, error)
}
