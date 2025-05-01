package repository_test

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Vyacheslav1557/tester/internal/contests/repository"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// setupTestDB creates a mocked sqlx.DB and sqlmock instance for testing.
func setupTestDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	return sqlxDB, mock
}

func TestRepository_CreateContest(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := repository.NewRepository(db)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()

		contest := models.Contest{
			Id:    1,
			Title: "Test Contest",
		}

		mock.ExpectQuery(repository.CreateContestQuery).
			WithArgs(contest.Title).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(contest.Id))

		id, err := repo.CreateContest(ctx, contest.Title)
		assert.NoError(t, err)
		assert.Equal(t, contest.Id, id)
	})
}

func TestRepository_GetContest(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := repository.NewRepository(db)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()

		contest := models.Contest{
			Id:        1,
			Title:     "Test Contest",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mock.ExpectQuery(repository.GetContestQuery).
			WithArgs(contest.Id).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "created_at", "updated_at"}).
				AddRow(contest.Id, contest.Title, contest.CreatedAt, contest.UpdatedAt))

		result, err := repo.GetContest(ctx, contest.Id)
		assert.NoError(t, err)
		assert.EqualExportedValues(t, &contest, result)
	})
}

func TestRepository_UpdateContest(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := repository.NewRepository(db)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()

		var contestId int32 = 1
		update := models.ContestUpdate{
			Title: sp("Updated Contest"),
		}

		mock.ExpectExec(repository.UpdateContestQuery).
			WithArgs(update.Title, contestId).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateContest(ctx, contestId, update)
		assert.NoError(t, err)
	})
}

func TestRepository_DeleteContest(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := repository.NewRepository(db)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()

		mock.ExpectExec(repository.DeleteContestQuery).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeleteContest(ctx, 1)
		assert.NoError(t, err)
	})
}

func sp(s string) *string {
	return &s
}
