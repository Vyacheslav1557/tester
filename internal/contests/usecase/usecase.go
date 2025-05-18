package usecase

import (
	"context"
	"github.com/Vyacheslav1557/tester/internal/contests"
	"github.com/Vyacheslav1557/tester/internal/models"
)

type UseCase struct {
	contestRepo contests.Repository
}

func NewContestUseCase(
	contestRepo contests.Repository,
) *UseCase {
	return &UseCase{
		contestRepo: contestRepo,
	}
}

func (uc *UseCase) CreateContest(ctx context.Context, title string) (int32, error) {
	return uc.contestRepo.CreateContest(ctx, title)
}

func (uc *UseCase) GetContest(ctx context.Context, id int32) (*models.Contest, error) {
	return uc.contestRepo.GetContest(ctx, id)
}

func (uc *UseCase) ListContests(ctx context.Context, filter models.ContestsFilter) (*models.ContestsList, error) {
	return uc.contestRepo.ListContests(ctx, filter)
}

func (uc *UseCase) UpdateContest(ctx context.Context, id int32, contestUpdate models.ContestUpdate) error {
	return uc.contestRepo.UpdateContest(ctx, id, contestUpdate)
}

func (uc *UseCase) DeleteContest(ctx context.Context, id int32) error {
	return uc.contestRepo.DeleteContest(ctx, id)
}

func (uc *UseCase) CreateContestProblem(ctx context.Context, contestId, problemId int32) error {
	return uc.contestRepo.CreateContestProblem(ctx, contestId, problemId)
}

func (uc *UseCase) GetContestProblem(ctx context.Context, contestId, problemId int32) (*models.ContestProblem, error) {
	return uc.contestRepo.GetContestProblem(ctx, contestId, problemId)
}

func (uc *UseCase) GetContestProblems(ctx context.Context, contestId int32) ([]*models.ContestProblemsListItem, error) {
	return uc.contestRepo.GetContestProblems(ctx, contestId)
}

func (uc *UseCase) DeleteContestProblem(ctx context.Context, contestId, problemId int32) error {
	return uc.contestRepo.DeleteContestProblem(ctx, contestId, problemId)
}

func (uc *UseCase) CreateParticipant(ctx context.Context, contestId, userId int32) error {
	return uc.contestRepo.CreateParticipant(ctx, contestId, userId)
}

func (uc *UseCase) IsParticipant(ctx context.Context, contestId int32, userId int32) (bool, error) {
	return uc.contestRepo.IsParticipant(ctx, contestId, userId)
}

func (uc *UseCase) DeleteParticipant(ctx context.Context, contestId, userId int32) error {
	return uc.contestRepo.DeleteParticipant(ctx, contestId, userId)
}

func (uc *UseCase) ListParticipants(ctx context.Context, filter models.ParticipantsFilter) (*models.UsersList, error) {
	return uc.contestRepo.ListParticipants(ctx, filter)
}

func (uc *UseCase) GetMonitor(ctx context.Context, contestId int32) (*models.Monitor, error) {
	return uc.contestRepo.GetMonitor(ctx, contestId)
}
