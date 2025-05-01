package usecase

import (
	"context"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/internal/sessions"
	"github.com/Vyacheslav1557/tester/internal/users"
	"github.com/Vyacheslav1557/tester/pkg"
	"github.com/google/uuid"
	"time"
)

type UseCase struct {
	usersUC    users.UseCase
	sessionsUC sessions.UseCase
}

func NewUseCase(usersUC users.UseCase, sessionsUC sessions.UseCase) *UseCase {
	return &UseCase{
		usersUC:    usersUC,
		sessionsUC: sessionsUC,
	}
}

func (uc *UseCase) Login(ctx context.Context, credentials *models.Credentials, device *models.Device) (*models.Session, error) {
	const op = "UseCase.Login"

	user, err := uc.usersUC.ReadUserByUsername(ctx, credentials.Username)
	if err != nil {
		return nil, err
	}

	if !user.IsSamePwd(credentials.Password) {
		return nil, pkg.Wrap(pkg.ErrNotFound, nil, op, "password mismatch")
	}

	session := &models.Session{
		Id:        uuid.NewString(),
		UserId:    user.Id,
		Role:      user.Role,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(40 * time.Minute),
		UserAgent: device.UseAgent,
		Ip:        device.Ip,
	}

	err = uc.sessionsUC.CreateSession(ctx, session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (uc *UseCase) Logout(ctx context.Context, sessionId string) error {
	return uc.sessionsUC.DeleteSession(ctx, sessionId)
}

func (uc *UseCase) Refresh(ctx context.Context, sessionId string) error {
	return uc.sessionsUC.UpdateSession(ctx, sessionId)
}

func (uc *UseCase) Terminate(ctx context.Context, userId int32) error {
	return uc.sessionsUC.DeleteAllSessions(ctx, userId)
}

func (uc *UseCase) ListSessions(ctx context.Context, userId int32) ([]*models.Session, error) {
	// TODO: implement me
	panic("implement me")
}
