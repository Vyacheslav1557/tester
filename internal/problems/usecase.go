package problems

import (
	"context"
	"github.com/Vyacheslav1557/tester/internal/models"
	"io"
)

type UseCase interface {
	CreateProblem(ctx context.Context, title string) (int32, error)
	GetProblemById(ctx context.Context, id int32) (*models.Problem, error)
	DeleteProblem(ctx context.Context, id int32) error
	ListProblems(ctx context.Context, filter models.ProblemsFilter) (*models.ProblemsList, error)
	UpdateProblem(ctx context.Context, id int32, problem *models.ProblemUpdate) error
	UploadProblem(ctx context.Context, id int32, r io.ReaderAt, size int64) error
	DownloadTestsArchive(ctx context.Context, id int32) (string, error)
}
