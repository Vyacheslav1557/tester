package usecase

import (
	"context"
	"github.com/Vyacheslav1557/tester/internal/models"
)

func (uc *ContestUseCase) GetSolution(ctx context.Context, id int32) (*models.Solution, error) {
	return uc.contestRepo.GetSolution(ctx, id)
}

func (uc *ContestUseCase) CreateSolution(ctx context.Context, creation *models.SolutionCreation) (int32, error) {
	participantId, err := uc.contestRepo.GetParticipantId2(ctx, creation.TaskId, creation.UserId)
	if err != nil {
		return 0, err
	}

	creation.ParticipantId = participantId

	return uc.contestRepo.CreateSolution(ctx, creation)
}

func (uc *ContestUseCase) ListSolutions(ctx context.Context, filter models.SolutionsFilter) (*models.SolutionsList, error) {
	return uc.contestRepo.ListSolutions(ctx, filter)
}

func (uc *ContestUseCase) GetBestSolutions(ctx context.Context, contestId int32, participantId int32) ([]*models.SolutionsListItem, error) {
	return uc.contestRepo.GetBestSolutions(ctx, contestId, participantId)
}
