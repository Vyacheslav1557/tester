package problems

import (
	"context"
	"database/sql"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/jmoiron/sqlx"
	"io"
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
	CreateProblem(ctx context.Context, q Querier, title string) (int32, error)
	GetProblemById(ctx context.Context, q Querier, id int32) (*models.Problem, error)
	DeleteProblem(ctx context.Context, q Querier, id int32) error
	ListProblems(ctx context.Context, q Querier, filter models.ProblemsFilter) (*models.ProblemsList, error)
	UpdateProblem(ctx context.Context, q Querier, id int32, heading *models.ProblemUpdate) error
}

type S3Repository interface {
	UploadTestsFile(ctx context.Context, id int32, reader io.Reader) (string, error)
	DownloadTestsFile(ctx context.Context, id int32) (io.ReadCloser, error)
}
