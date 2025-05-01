package repository_test

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Vyacheslav1557/tester/internal/contests/repository"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRepository_CreateTask(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := repository.NewRepository(db)

	t.Run("success", func(t *testing.T) {
		var (
			expectedId int32 = 1
			problemId  int32 = 2
			contestId  int32 = 3
		)
		ctx := context.Background()

		mock.ExpectQuery(repository.CreateTaskQuery).
			WithArgs(problemId, contestId).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedId))

		id, err := repo.CreateTask(ctx, contestId, problemId)
		assert.NoError(t, err)
		assert.Equal(t, expectedId, id)
	})
}

func TestRepository_DeleteTask(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := repository.NewRepository(db)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()

		mock.ExpectExec(repository.DeleteTaskQuery).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeleteTask(ctx, 1)
		assert.NoError(t, err)
	})
}
