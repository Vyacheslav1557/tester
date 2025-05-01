package users

import (
	"context"
	"database/sql"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/jmoiron/sqlx"
)

type Querier interface {
	Rebind(query string) string
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

type Tx interface {
	Querier
	Commit() error
	Rollback() error
}

type Repository interface {
	BeginTx(ctx context.Context) (Tx, error)
	DB() Querier
	CreateUser(ctx context.Context, q Querier, user *models.UserCreation) (int32, error)
	ReadUserByUsername(ctx context.Context, q Querier, username string) (*models.User, error)
	ReadUserById(ctx context.Context, q Querier, id int32) (*models.User, error)
	UpdateUser(ctx context.Context, q Querier, id int32, update *models.UserUpdate) error
	DeleteUser(ctx context.Context, q Querier, id int32) error
	ListUsers(ctx context.Context, q Querier, filters models.UsersListFilters) (*models.UsersList, error)
}
