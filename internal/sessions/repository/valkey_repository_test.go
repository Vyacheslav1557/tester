package repository_test

import (
	"context"
	"fmt"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/internal/sessions/repository"
	"github.com/Vyacheslav1557/tester/pkg"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/valkey-io/valkey-go"
	"github.com/valkey-io/valkey-go/mock"
	"go.uber.org/mock/gomock"
	"strconv"
	"testing"
	"time"
)

func TestValkeyRepository_CreateSession(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewClient(ctrl)
	sessionRepo := repository.NewValkeyRepository(client)

	t.Run("success", func(t *testing.T) {
		session := &models.Session{
			Id:        uuid.NewString(),
			UserId:    1,
			Role:      models.RoleAdmin,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(repository.SessionLifetime),
			UserAgent: "Mozilla/5.0",
			Ip:        "127.0.0.1",
		}

		matcher := mock.MatchFn(func(cmd []string) bool {
			if cmd[0] != "SET" {
				return false
			}
			if cmd[1] != session.Key() {
				return false
			}
			if cmd[3] != "EXAT" {
				return false
			}
			if cmd[4] != strconv.FormatInt(session.ExpiresAt.Unix(), 10) {
				return false
			}
			return true
		})

		ctx := context.Background()
		client.EXPECT().Do(ctx, matcher)
		err := sessionRepo.CreateSession(ctx, session)
		require.NoError(t, err)
	})
}

func TestValkeyRepository_ReadSession(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewClient(ctrl)
	sessionRepo := repository.NewValkeyRepository(client)

	t.Run("success", func(t *testing.T) {
		session := &models.Session{
			Id:        uuid.NewString(),
			UserId:    1,
			Role:      models.RoleAdmin,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(repository.SessionLifetime),
			UserAgent: "Mozilla/5.0",
			Ip:        "127.0.0.1",
		}

		matcher := mock.MatchFn(func(cmd []string) bool {
			fmt.Println(cmd)

			if cmd[0] != "EVALSHA" {
				return false
			}
			if cmd[2] != "0" {
				return false
			}
			if cmd[3] != fmt.Sprintf("userid:*:sessionid:%s", session.SessionIdHash()) {
				return false
			}
			return true
		})

		d, err := session.JSON()
		require.NoError(t, err)
		ctx := context.Background()
		client.EXPECT().Do(ctx, matcher).Return(mock.Result(mock.ValkeyString(string(d))))
		res, err := sessionRepo.ReadSession(ctx, session.Id)
		require.NoError(t, err)
		fmt.Println(res.CreatedAt.Unix(), res.ExpiresAt.UnixNano())
		fmt.Println(session.CreatedAt.Unix(), session.ExpiresAt.UnixNano())
		require.EqualExportedValues(t, session, res)
	})

	t.Run("not found", func(t *testing.T) {
		session := &models.Session{
			Id: uuid.NewString(),
		}

		matcher := mock.MatchFn(func(cmd []string) bool {
			if cmd[0] != "EVALSHA" {
				return false
			}
			if cmd[2] != "0" {
				return false
			}
			if cmd[3] != fmt.Sprintf("userid:*:sessionid:%s", session.SessionIdHash()) {
				return false
			}
			return true
		})

		ctx := context.Background()
		client.EXPECT().Do(ctx, matcher).Return(mock.ErrorResult(valkey.Nil))
		res, err := sessionRepo.ReadSession(ctx, session.Id)
		require.ErrorIs(t, err, pkg.ErrNotFound)
		require.ErrorIs(t, err, valkey.Nil)
		require.Empty(t, res)
	})
}

func TestValkeyRepository_UpdateSession(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewClient(ctrl)
	sessionRepo := repository.NewValkeyRepository(client)

	t.Run("success", func(t *testing.T) {
		session := &models.Session{
			Id: uuid.NewString(),
		}

		matcher := mock.MatchFn(func(cmd []string) bool {
			if cmd[0] != "EVALSHA" {
				return false
			}
			if cmd[2] != "0" {
				return false
			}
			if cmd[3] != fmt.Sprintf("userid:*:sessionid:%s", session.SessionIdHash()) {
				return false
			}
			return true
		})

		ctx := context.Background()
		client.EXPECT().Do(ctx, matcher)
		err := sessionRepo.UpdateSession(ctx, session.Id)
		require.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		session := &models.Session{
			Id: uuid.NewString(),
		}

		matcher := mock.MatchFn(func(cmd []string) bool {
			if cmd[0] != "EVALSHA" {
				return false
			}
			if cmd[2] != "0" {
				return false
			}
			if cmd[3] != fmt.Sprintf("userid:*:sessionid:%s", session.SessionIdHash()) {
				return false
			}
			return true
		})

		ctx := context.Background()
		client.EXPECT().Do(ctx, matcher).Return(mock.ErrorResult(valkey.Nil))
		err := sessionRepo.UpdateSession(ctx, session.Id)
		require.ErrorIs(t, err, pkg.ErrNotFound)
		require.ErrorIs(t, err, valkey.Nil)
	})
}

func TestValkeyRepository_DeleteSession(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewClient(ctrl)
	sessionRepo := repository.NewValkeyRepository(client)

	t.Run("success", func(t *testing.T) {
		session := &models.Session{
			Id: uuid.NewString(),
		}

		matcher := mock.MatchFn(func(cmd []string) bool {
			if cmd[0] != "EVALSHA" {
				return false
			}
			if cmd[2] != "0" {
				return false
			}
			if cmd[3] != fmt.Sprintf("userid:*:sessionid:%s", session.SessionIdHash()) {
				return false
			}
			return true
		})

		ctx := context.Background()
		client.EXPECT().Do(ctx, matcher)
		err := sessionRepo.DeleteSession(ctx, session.Id)
		require.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		session := &models.Session{
			Id: uuid.NewString(),
		}

		matcher := mock.MatchFn(func(cmd []string) bool {
			if cmd[0] != "EVALSHA" {
				return false
			}
			if cmd[2] != "0" {
				return false
			}
			if cmd[3] != fmt.Sprintf("userid:*:sessionid:%s", session.SessionIdHash()) {
				return false
			}
			return true
		})

		ctx := context.Background()
		client.EXPECT().Do(ctx, matcher).Return(mock.ErrorResult(valkey.Nil))
		err := sessionRepo.DeleteSession(ctx, session.Id)
		require.ErrorIs(t, err, pkg.ErrNotFound)
		require.ErrorIs(t, err, valkey.Nil)
	})
}

func TestValkeyRepository_DeleteAllSessions(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewClient(ctrl)
	sessionRepo := repository.NewValkeyRepository(client)

	t.Run("success", func(t *testing.T) {
		session := &models.Session{
			UserId: 1,
		}

		matcher := mock.MatchFn(func(cmd []string) bool {
			fmt.Println(cmd)

			if cmd[0] != "EVALSHA" {
				return false
			}
			if cmd[2] != "0" {
				return false
			}
			if cmd[3] != fmt.Sprintf("userid:%s:sessionid:*", session.UserIdHash()) {
				return false
			}
			return true
		})

		ctx := context.Background()
		client.EXPECT().Do(ctx, matcher)
		err := sessionRepo.DeleteAllSessions(ctx, session.UserId)
		require.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		session := &models.Session{
			UserId: 1,
		}

		matcher := mock.MatchFn(func(cmd []string) bool {
			if cmd[0] != "EVALSHA" {
				return false
			}
			if cmd[2] != "0" {
				return false
			}
			if cmd[3] != fmt.Sprintf("userid:%s:sessionid:*", session.UserIdHash()) {
				return false
			}
			return true
		})

		ctx := context.Background()
		client.EXPECT().Do(ctx, matcher).Return(mock.ErrorResult(valkey.Nil))
		err := sessionRepo.DeleteAllSessions(ctx, session.UserId)
		require.ErrorIs(t, err, pkg.ErrNotFound)
		require.ErrorIs(t, err, valkey.Nil)
	})
}
