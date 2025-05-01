package repository

import (
	"context"
	"fmt"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/pkg"
	"github.com/valkey-io/valkey-go"
	"strconv"
	"time"
)

type ValkeyRepository struct {
	db valkey.Client
}

func NewValkeyRepository(db valkey.Client) *ValkeyRepository {
	return &ValkeyRepository{
		db: db,
	}
}

const SessionLifetime = time.Minute * 40

func (r *ValkeyRepository) CreateSession(ctx context.Context, session *models.Session) error {
	const op = "ValkeyRepository.CreateSession"

	data, err := session.JSON()
	if err != nil {
		return pkg.Wrap(pkg.ErrInternal, err, op, "cannot marshal session")
	}

	resp := r.db.Do(ctx, r.db.
		B().Set().
		Key(session.Key()).
		Value(string(data)).
		Exat(session.ExpiresAt).
		Build(),
	)

	err = resp.Error()
	if err != nil {
		if valkey.IsValkeyNil(err) {
			return pkg.Wrap(pkg.ErrInternal, err, op, "nil response")
		}
		return pkg.Wrap(pkg.ErrUnhandled, err, op, "unhandled valkey error")
	}

	return nil
}

const (
	readSessionScript = `local result = redis.call('SCAN', 0, 'MATCH', ARGV[1])
if #result[2] == 0 then
	return nil
else
	return redis.call('GET', result[2][1])
end`
)

func (r *ValkeyRepository) ReadSession(ctx context.Context, sessionId string) (*models.Session, error) {
	const op = "ValkeyRepository.ReadSession"

	sessionIdHash := (&models.Session{Id: sessionId}).SessionIdHash()

	resp := valkey.NewLuaScript(readSessionScript).Exec(
		ctx,
		r.db,
		nil,
		[]string{fmt.Sprintf("userid:*:sessionid:%s", sessionIdHash)},
	)

	if err := resp.Error(); err != nil {
		if valkey.IsValkeyNil(err) {
			return nil, pkg.Wrap(pkg.ErrNotFound, err, op, "reading session")
		}
		return nil, pkg.Wrap(pkg.ErrUnhandled, err, op, "unhandled valkey error")
	}

	session := &models.Session{}

	err := resp.DecodeJSON(session)
	if err != nil {
		return nil, pkg.Wrap(pkg.ErrInternal, err, op, "session storage corrupted")
	}

	return session, nil
}

const (
	updateSessionScript = `local result = redis.call('SCAN', 0, 'MATCH', ARGV[1])   
return #result[2] > 0 and redis.call('EXPIRE', result[2][1], ARGV[2]) == 1`
)

var (
	sessionLifetimeString = strconv.Itoa(int(SessionLifetime.Seconds()))
)

func (r *ValkeyRepository) UpdateSession(ctx context.Context, sessionId string) error {
	const op = "ValkeyRepository.UpdateSession"

	sessionIdHash := (&models.Session{Id: sessionId}).SessionIdHash()

	resp := valkey.NewLuaScript(updateSessionScript).Exec(
		ctx,
		r.db,
		nil,
		[]string{fmt.Sprintf("userid:*:sessionid:%s", sessionIdHash), sessionLifetimeString},
	)

	err := resp.Error()
	if err != nil {
		if valkey.IsValkeyNil(err) {
			return pkg.Wrap(pkg.ErrNotFound, err, op, "nil response")
		}
		return pkg.Wrap(pkg.ErrUnhandled, err, op, "unhandled valkey error")
	}

	return nil
}

const deleteSessionScript = `local result = redis.call('SCAN', 0, 'MATCH', ARGV[1])   
return #result[2] > 0 and redis.call('DEL', result[2][1]) == 1`

func (r *ValkeyRepository) DeleteSession(ctx context.Context, sessionId string) error {
	const op = "ValkeyRepository.DeleteSession"

	sessionIdHash := (&models.Session{Id: sessionId}).SessionIdHash()

	resp := valkey.NewLuaScript(deleteSessionScript).Exec(
		ctx,
		r.db,
		nil,
		[]string{fmt.Sprintf("userid:*:sessionid:%s", sessionIdHash)},
	)

	err := resp.Error()
	if err != nil {
		if valkey.IsValkeyNil(err) {
			return pkg.Wrap(pkg.ErrNotFound, err, op, "nil response")
		}
		return pkg.Wrap(pkg.ErrUnhandled, err, op, "unhandled valkey error")
	}

	return nil
}

const (
	deleteUserSessionsScript = `local cursor = 0
local dels = 0
repeat
    local result = redis.call('SCAN', cursor, 'MATCH', ARGV[1])
    for _,key in ipairs(result[2]) do
        redis.call('DEL', key)
        dels = dels + 1
    end
    cursor = tonumber(result[1])
until cursor == 0
return dels`
)

func (r *ValkeyRepository) DeleteAllSessions(ctx context.Context, userId int32) error {
	const op = "ValkeyRepository.DeleteAllSessions"

	userIdHash := (&models.Session{UserId: userId}).UserIdHash()

	resp := valkey.NewLuaScript(deleteUserSessionsScript).Exec(
		ctx,
		r.db,
		nil,
		[]string{fmt.Sprintf("userid:%s:sessionid:*", userIdHash)},
	)

	err := resp.Error()
	if err != nil {
		if valkey.IsValkeyNil(err) {
			return pkg.Wrap(pkg.ErrNotFound, err, op, "nil response")
		}
		return pkg.Wrap(pkg.ErrUnhandled, err, op, "unhandled valkey error")
	}

	return nil
}
