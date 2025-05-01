package repository_test

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Vyacheslav1557/tester/internal/contests/repository"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRepository_CreateParticipant(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := repository.NewRepository(db)

	t.Run("success", func(t *testing.T) {
		var (
			expectedId int32 = 1
			userId     int32 = 2
			contestId  int32 = 3
		)
		ctx := context.Background()

		mock.ExpectQuery(repository.CreateParticipantQuery).
			WithArgs(userId, contestId).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedId))

		id, err := repo.CreateParticipant(ctx, contestId, userId)
		assert.NoError(t, err)
		assert.Equal(t, expectedId, id)
	})
}

func TestRepository_DeleteParticipant(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := repository.NewRepository(db)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		var participantId int32 = 1

		mock.ExpectExec(repository.DeleteParticipantQuery).
			WithArgs(participantId).WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeleteParticipant(ctx, participantId)
		assert.NoError(t, err)
	})
}
