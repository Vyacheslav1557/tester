package users

import (
	"context"
	"github.com/Vyacheslav1557/tester/internal/models"
)

type UseCase interface {
	CreateUser(ctx context.Context, user *models.UserCreation) (int32, error)
	ReadUserById(ctx context.Context, id int32) (*models.User, error)
	ReadUserByUsername(ctx context.Context, username string) (*models.User, error)
	UpdateUser(ctx context.Context, id int32, update *models.UserUpdate) error
	DeleteUser(ctx context.Context, id int32) error
	ListUsers(ctx context.Context, filters models.UsersListFilters) (*models.UsersList, error)
}
