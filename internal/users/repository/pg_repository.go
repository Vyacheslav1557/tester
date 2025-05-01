package repository

import (
	"context"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/internal/users"
	"github.com/Vyacheslav1557/tester/pkg"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	_db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		_db: db,
	}
}

func (r *Repository) BeginTx(ctx context.Context) (users.Tx, error) {
	tx, err := r._db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (r *Repository) DB() users.Querier {
	return r._db
}

const CreateUserQuery = `
INSERT INTO users
    (username, hashed_pwd, role)
VALUES ($1, $2, $3)
RETURNING id
`

func (r *Repository) CreateUser(ctx context.Context, q users.Querier, user *models.UserCreation) (int32, error) {
	const op = "Caller.CreateUser"

	rows, err := q.QueryxContext(
		ctx,
		CreateUserQuery,
		user.Username,
		user.Password,
		user.Role,
	)
	if err != nil {
		return 0, pkg.HandlePgErr(err, op)
	}

	defer rows.Close()
	var id int32
	rows.Next()
	err = rows.Scan(&id)
	if err != nil {
		return 0, pkg.HandlePgErr(err, op)
	}

	return id, nil
}

const ReadUserByUsernameQuery = "SELECT * from users WHERE username=$1 LIMIT 1"

func (r *Repository) ReadUserByUsername(ctx context.Context, q users.Querier, username string) (*models.User, error) {
	const op = "Caller.ReadUserByUsername"

	var user models.User
	err := q.GetContext(ctx, &user, ReadUserByUsernameQuery, username)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}
	return &user, nil
}

const ReadUserByIdQuery = "SELECT * from users WHERE id=$1 LIMIT 1"

func (r *Repository) ReadUserById(ctx context.Context, q users.Querier, id int32) (*models.User, error) {
	const op = "Caller.ReadUserById"

	var user models.User
	err := q.GetContext(ctx, &user, ReadUserByIdQuery, id)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}
	return &user, nil
}

const UpdateUserQuery = `
UPDATE users
SET username        = COALESCE($1, username),
	role            = COALESCE($2, role)
WHERE id = $3
`

func (r *Repository) UpdateUser(ctx context.Context, q users.Querier, id int32, update *models.UserUpdate) error {
	const op = "Caller.UpdateUser"

	_, err := q.ExecContext(
		ctx,
		UpdateUserQuery,
		update.Username,
		update.Role,
		id,
	)

	if err != nil {
		return pkg.HandlePgErr(err, op)
	}
	return nil
}

const DeleteUserQuery = "DELETE FROM users WHERE id = $1"

func (r *Repository) DeleteUser(ctx context.Context, q users.Querier, id int32) error {
	const op = "Caller.DeleteUser"

	_, err := q.ExecContext(ctx, DeleteUserQuery, id)
	if err != nil {
		return pkg.HandlePgErr(err, op)
	}

	return nil
}

const (
	ListUsersQuery  = "SELECT * FROM users LIMIT $1 OFFSET $2"
	CountUsersQuery = "SELECT COUNT(*) FROM users"
)

func (r *Repository) ListUsers(ctx context.Context, q users.Querier, filters models.UsersListFilters) (*models.UsersList, error) {
	const op = "Caller.ListUsers"

	list := make([]*models.User, 0)
	err := q.SelectContext(ctx, &list, ListUsersQuery, filters.PageSize, filters.Offset())
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	var count int32
	err = q.GetContext(ctx, &count, CountUsersQuery)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	return &models.UsersList{
		Users: list,
		Pagination: models.Pagination{
			Total: models.Total(count, filters.PageSize),
			Page:  filters.Page,
		},
	}, nil
}
