package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/internal/users/repository"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

// setupTestDB creates a mocked sqlx.DB and sqlmock instance for runner.
func setupTestDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	return sqlxDB, mock
}

func TestRepository_CreateUser(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := repository.NewRepository(db)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()

		var expectedId int32 = 1
		user := &models.UserCreation{
			Username: "testuser",
			Password: "hashed-password",
			Role:     models.RoleAdmin,
		}

		mock.ExpectQuery(repository.CreateUserQuery).
			WithArgs(user.Username, sqlmock.AnyArg(), user.Role).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedId))

		id, err := repo.CreateUser(ctx, db, user)
		assert.NoError(t, err)
		assert.Equal(t, expectedId, id)
	})
}

func TestRepository_ReadUserByUsername(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := repository.NewRepository(db)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()

		expected := &models.User{
			Id:             1,
			Username:       "testuser",
			HashedPassword: "hashed-password",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			Role:           models.RoleAdmin,
		}

		columns := []string{
			"id",
			"username",
			"hashed_pwd",
			"created_at",
			"updated_at",
			"role",
		}

		rows := sqlmock.NewRows(columns).AddRow(
			expected.Id,
			expected.Username,
			expected.HashedPassword,
			expected.CreatedAt,
			expected.UpdatedAt,
			expected.Role,
		)

		mock.ExpectQuery(repository.ReadUserByUsernameQuery).WithArgs(expected.Username).WillReturnRows(rows)

		user, err := repo.ReadUserByUsername(ctx, db, expected.Username)
		assert.NoError(t, err)
		assert.Equal(t, expected, user)
	})

	t.Run("not found", func(t *testing.T) {
		ctx := context.Background()

		username := "testuser"

		mock.ExpectQuery(repository.ReadUserByUsernameQuery).WithArgs(username).WillReturnError(sql.ErrNoRows)

		user, err := repo.ReadUserByUsername(ctx, db, username)
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestRepository_ReadUserById(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := repository.NewRepository(db)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()

		expected := &models.User{
			Id:       1,
			Username: "testuser",
			Role:     models.RoleAdmin,
		}

		mock.ExpectQuery(repository.ReadUserByIdQuery).
			WithArgs(expected.Id).
			WillReturnRows(sqlmock.NewRows([]string{"id", "username", "role"}).
				AddRow(expected.Id, expected.Username, expected.Role))

		user, err := repo.ReadUserById(ctx, db, expected.Id)
		assert.NoError(t, err)
		assert.Equal(t, expected, user)
	})

	t.Run("not found", func(t *testing.T) {
		ctx := context.Background()

		userID := int32(1)

		mock.ExpectQuery(repository.ReadUserByIdQuery).WithArgs(userID).WillReturnError(sql.ErrNoRows)

		user, err := repo.ReadUserById(ctx, db, userID)
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestRepository_UpdateUser(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := repository.NewRepository(db)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()

		userID := int32(1)
		username := "testuser"
		role := models.RoleStudent
		update := &models.UserUpdate{
			Username: &username,
			Role:     &role,
		}

		mock.ExpectExec(repository.UpdateUserQuery).
			WithArgs(update.Username, update.Role, userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateUser(ctx, db, userID, update)
		assert.NoError(t, err)
	})
}

func TestRepository_DeleteUser(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := repository.NewRepository(db)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		userID := int32(1)

		mock.ExpectExec(repository.DeleteUserQuery).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeleteUser(ctx, db, userID)
		assert.NoError(t, err)
	})
}

//func TestRepository_ListUsers(t *testing.T) {
//	db, mock := setupTestDB(t)
//	defer db.Close()
//
//	repo := repository.NewRepository(db)
//
//	t.Run("success", func(t *testing.T) {
//		ctx := context.Background()
//
//		filters := models.UsersListFilters{
//			Page:     1,
//			PageSize: 10,
//		}
//		expectedUsers := []*models.User{
//			{Id: 1, Username: "user1", Role: models.RoleAdmin},
//			{Id: 2, Username: "user2", Role: models.RoleStudent},
//		}
//		totalCount := int32(2)
//
//		mock.ExpectQuery(repository.ListUsersQuery).
//			WithArgs(filters.PageSize, filters.Offset()).
//			WillReturnRows(sqlmock.NewRows([]string{"id", "username", "role"}).
//				AddRow(expectedUsers[0].Id, expectedUsers[0].Username, expectedUsers[0].Role).
//				AddRow(expectedUsers[1].Id, expectedUsers[1].Username, expectedUsers[1].Role))
//
//		mock.ExpectQuery(repository.CountUsersQuery).
//			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(totalCount))
//
//		result, err := repo.ListUsers(ctx, db, filters)
//		assert.NoError(t, err)
//		assert.Equal(t, expectedUsers, result.Users)
//		assert.Equal(t, models.Pagination{Total: 1, Page: 1}, result.Pagination)
//	})
//}
