package solutions

import (
	"context"
	"github.com/Vyacheslav1557/tester/internal/models"
)

type UseCase interface {
	GetSolution(ctx context.Context, id int32) (*models.Solution, error)
	CreateSolution(ctx context.Context, creation *models.SolutionCreation) (int32, error)
	UpdateSolution(ctx context.Context, id int32, update *models.SolutionUpdate) error
	ListSolutions(ctx context.Context, filter models.SolutionsFilter) (*models.SolutionsList, error)
}
