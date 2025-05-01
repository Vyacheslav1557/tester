package pkg

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"net/http"
)

var (
	NoPermission       = errors.New("no permission")
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrUnhandled       = errors.New("unhandled")
	ErrNotFound        = errors.New("not found")
	ErrBadInput        = errors.New("bad input")
	ErrInternal        = errors.New("internal")
)

func Wrap(basic error, err error, op string, msg string) error {
	return errors.Join(basic, err, fmt.Errorf("during %s: %s", op, msg))
}

func ToREST(err error) int {
	switch {
	case errors.Is(err, ErrUnauthenticated):
		return http.StatusUnauthorized
	case errors.Is(err, ErrBadInput):
		return http.StatusBadRequest
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrInternal):
		return http.StatusInternalServerError
	case errors.Is(err, NoPermission):
		return http.StatusForbidden
	}

	return http.StatusInternalServerError
}

func HandlePgErr(err error, op string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return Wrap(ErrBadInput, err, op, pgErr.Message)
		}
		if pgerrcode.IsNoData(pgErr.Code) {
			return Wrap(ErrNotFound, err, op, pgErr.Message)
		}
	}

	if errors.Is(err, sql.ErrNoRows) {
		return Wrap(ErrNotFound, err, op, "no rows found")
	}

	return Wrap(ErrUnhandled, err, op, "unexpected error")
}
