package usecase

import (
	"context"
	"errors"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/internal/sessions"
	"github.com/Vyacheslav1557/tester/internal/users"
	"github.com/Vyacheslav1557/tester/pkg"
)

type UsersUC struct {
	sessionRepo sessions.ValkeyRepository
	usersRepo   users.Repository
}

func NewUseCase(sessionRepo sessions.ValkeyRepository, usersRepo users.Repository) *UsersUC {
	return &UsersUC{
		sessionRepo: sessionRepo,
		usersRepo:   usersRepo,
	}
}

func (u *UsersUC) CreateUser(ctx context.Context, user *models.UserCreation) (int32, error) {
	const op = "UseCase.CreateUser"

	err := user.HashPassword()
	if err != nil {
		return 0, pkg.Wrap(pkg.ErrBadInput, err, op, "bad password")
	}

	id, err := u.usersRepo.CreateUser(ctx, u.usersRepo.DB(), user)
	if err != nil {
		return 0, pkg.Wrap(nil, err, op, "can't create user")
	}

	return id, nil
}

func (u *UsersUC) ListUsers(ctx context.Context, filters models.UsersListFilters) (*models.UsersList, error) {
	const op = "UseCase.ListUsers"

	usersList, err := u.usersRepo.ListUsers(ctx, u.usersRepo.DB(), filters)
	if err != nil {
		return nil, pkg.Wrap(nil, err, op, "can't list users")
	}

	return usersList, nil
}

func (u *UsersUC) UpdateUser(ctx context.Context, id int32, update *models.UserUpdate) error {
	const op = "UseCase.UpdateUser"

	tx, err := u.usersRepo.BeginTx(ctx)
	if err != nil {
		return pkg.Wrap(nil, err, op, "cannot start transaction")
	}

	err = u.usersRepo.UpdateUser(ctx, tx, id, update)
	if err != nil {
		return pkg.Wrap(nil, errors.Join(err, tx.Rollback()), op, "cannot update user")
	}
	err = u.sessionRepo.DeleteAllSessions(ctx, id)
	if err != nil {
		return pkg.Wrap(nil, errors.Join(err, tx.Rollback()), op, "cannot delete all sessions")
	}
	err = tx.Commit()
	if err != nil {
		return pkg.Wrap(nil, err, op, "cannot commit transaction")
	}

	return nil
}

// ReadUserByUsername is for login only. There are no permission checks! DO NOT USE IT AS AN ENDPOINT RESPONSE!
func (u *UsersUC) ReadUserByUsername(ctx context.Context, username string) (*models.User, error) {
	const op = "UseCase.ReadUserByUsername"

	user, err := u.usersRepo.ReadUserByUsername(ctx, u.usersRepo.DB(), username)
	if err != nil {
		return nil, pkg.Wrap(nil, err, op, "can't read user by username")
	}
	return user, nil
}

func (u *UsersUC) ReadUserById(ctx context.Context, id int32) (*models.User, error) {
	const op = "UseCase.ReadUserById"

	user, err := u.usersRepo.ReadUserById(ctx, u.usersRepo.DB(), id)
	if err != nil {
		return nil, pkg.Wrap(nil, err, op, "can't read user by id")
	}
	return user, nil
}

func (u *UsersUC) DeleteUser(ctx context.Context, id int32) error {
	const op = "UseCase.DeleteUser"

	tx, err := u.usersRepo.BeginTx(ctx)
	if err != nil {
		return pkg.Wrap(nil, err, op, "cannot start transaction")
	}

	err = u.usersRepo.DeleteUser(ctx, tx, id)
	if err != nil {
		return pkg.Wrap(nil, errors.Join(err, tx.Rollback()), op, "cannot delete user")
	}

	err = u.sessionRepo.DeleteAllSessions(ctx, id)
	if err != nil {
		return pkg.Wrap(nil, errors.Join(err, tx.Rollback()), op, "cannot delete all sessions")
	}
	err = tx.Commit()
	if err != nil {
		return pkg.Wrap(nil, err, op, "cannot commit transaction")
	}

	return nil
}

/*
func ValidEmail(str string) error {
	emailAddress, err := mail.ParseAddress(str)
	if err != nil || emailAddress.Address != str {
		return errors.New("invalid email")
	}
	return nil
}

func ValidUsername(str string) error {
	if len(str) < 5 {
		return errors.New("too short username")
	}
	if len(str) > 70 {
		return errors.New("too long username")
	}
	if err := ValidEmail(str); err == nil {
		return errors.New("username cannot be an email")
	}
	return nil
}

func ValidPassword(str string) error {
	if len(str) < 5 {
		return errors.New("too short password")
	}
	if len(str) > 70 {
		return errors.New("too long password")
	}
	return nil
}

func ValidRole(role models.Role) error {
	switch role {
	case models.RoleAdmin:
		return nil
	case models.RoleTeacher:
		return nil
	case models.RoleStudent:
		return nil
	}
	return errors.New("invalid role")
}
*/
