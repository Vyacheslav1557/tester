package usecase

import (
	"context"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/internal/solutions"
)

type UseCase struct {
	solutionsRepo solutions.Repository
}

func NewUseCase(solutionsRepo solutions.Repository) *UseCase {
	return &UseCase{
		solutionsRepo: solutionsRepo,
	}
}

func (uc *UseCase) GetSolution(ctx context.Context, id int32) (*models.Solution, error) {
	return uc.solutionsRepo.GetSolution(ctx, id)
}

func (uc *UseCase) CreateSolution(ctx context.Context, creation *models.SolutionCreation) (int32, error) {
	return uc.solutionsRepo.CreateSolution(ctx, creation)
}

func (uc *UseCase) UpdateSolution(ctx context.Context, id int32, update *models.SolutionUpdate) error {
	return uc.solutionsRepo.UpdateSolution(ctx, id, update)
}

func (uc *UseCase) ListSolutions(ctx context.Context, filter models.SolutionsFilter) (*models.SolutionsList, error) {
	return uc.solutionsRepo.ListSolutions(ctx, filter)
}
