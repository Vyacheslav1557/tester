package usecase

import (
	"context"
	"github.com/Vyacheslav1557/tester/internal/contests"
	"github.com/Vyacheslav1557/tester/internal/models"
)

type ContestUseCase struct {
	contestRepo contests.Repository
}

func NewContestUseCase(
	contestRepo contests.Repository,
) *ContestUseCase {
	return &ContestUseCase{
		contestRepo: contestRepo,
	}
}

func (uc *ContestUseCase) CreateContest(ctx context.Context, title string) (int32, error) {
	return uc.contestRepo.CreateContest(ctx, title)
}

func (uc *ContestUseCase) GetContest(ctx context.Context, id int32) (*models.Contest, error) {
	return uc.contestRepo.GetContest(ctx, id)
}

func (uc *ContestUseCase) UpdateContest(ctx context.Context, id int32, contestUpdate models.ContestUpdate) error {
	return uc.contestRepo.UpdateContest(ctx, id, contestUpdate)
}

func (uc *ContestUseCase) DeleteContest(ctx context.Context, id int32) error {
	return uc.contestRepo.DeleteContest(ctx, id)
}

func (uc *ContestUseCase) ListContests(ctx context.Context, filter models.ContestsFilter) (*models.ContestsList, error) {
	return uc.contestRepo.ListContests(ctx, filter)
}
