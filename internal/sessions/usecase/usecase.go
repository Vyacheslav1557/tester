package usecase

import (
	"context"
	"github.com/Vyacheslav1557/tester/config"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/internal/sessions"
	"github.com/Vyacheslav1557/tester/pkg"
)

type SessionsUC struct {
	sessionsRepo sessions.ValkeyRepository
	cfg          config.Config
}

func NewUseCase(
	sessionRepo sessions.ValkeyRepository,
	cfg config.Config,
) *SessionsUC {
	return &SessionsUC{
		sessionsRepo: sessionRepo,
		cfg:          cfg,
	}
}

// CreateSession is for login only. There are no permission checks! DO NOT USE IT AS AN ENDPOINT RESPONSE!
func (u *SessionsUC) CreateSession(ctx context.Context, creation *models.Session) error {
	const op = "UseCase.CreateSession"

	err := u.sessionsRepo.CreateSession(ctx, creation)
	if err != nil {
		return pkg.Wrap(nil, err, op, "cannot create session")
	}

	return nil
}

// ReadSession is for internal use only. There are no permission checks! DO NOT USE IT AS AN ENDPOINT RESPONSE!
func (u *SessionsUC) ReadSession(ctx context.Context, sessionId string) (*models.Session, error) {
	const op = "UseCase.ReadSession"

	session, err := u.sessionsRepo.ReadSession(ctx, sessionId)
	if err != nil {
		return nil, pkg.Wrap(nil, err, op, "cannot read session")
	}
	return session, nil
}

func (u *SessionsUC) UpdateSession(ctx context.Context, sessionId string) error {
	const op = "UseCase.UpdateSession"

	err := u.sessionsRepo.UpdateSession(ctx, sessionId)
	if err != nil {
		return pkg.Wrap(nil, err, op, "cannot update session")
	}
	return nil
}

func (u *SessionsUC) DeleteSession(ctx context.Context, sessionId string) error {
	const op = "UseCase.DeleteSession"

	err := u.sessionsRepo.DeleteSession(ctx, sessionId)
	if err != nil {
		return pkg.Wrap(nil, err, op, "cannot delete session")
	}
	return nil
}

func (u *SessionsUC) DeleteAllSessions(ctx context.Context, userId int32) error {
	const op = "UseCase.DeleteAllSessions"

	err := u.sessionsRepo.DeleteAllSessions(ctx, userId)
	if err != nil {
		return pkg.Wrap(nil, err, op, "cannot delete all sessions")
	}

	return nil
}
