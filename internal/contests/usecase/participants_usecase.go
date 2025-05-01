package usecase

import (
	"context"
	"github.com/Vyacheslav1557/tester/internal/models"
)

func (uc *ContestUseCase) GetParticipantId(ctx context.Context, contestId int32, userId int32) (int32, error) {
	return uc.contestRepo.GetParticipantId(ctx, contestId, userId)
}

func (uc *ContestUseCase) GetParticipantId2(ctx context.Context, taskId, userId int32) (int32, error) {
	return uc.contestRepo.GetParticipantId2(ctx, taskId, userId)
}

func (uc *ContestUseCase) GetParticipantId3(ctx context.Context, solutionId int32) (int32, error) {
	return uc.contestRepo.GetParticipantId3(ctx, solutionId)
}

func (uc *ContestUseCase) CreateParticipant(ctx context.Context, contestId int32, userId int32) (id int32, err error) {
	return uc.contestRepo.CreateParticipant(ctx, contestId, userId)
}

func (uc *ContestUseCase) DeleteParticipant(ctx context.Context, participantId int32) error {
	return uc.contestRepo.DeleteParticipant(ctx, participantId)
}

func (uc *ContestUseCase) ListParticipants(ctx context.Context, filter models.ParticipantsFilter) (*models.ParticipantsList, error) {
	return uc.contestRepo.ListParticipants(ctx, filter)
}

func (uc *ContestUseCase) UpdateParticipant(ctx context.Context, id int32, participantUpdate models.ParticipantUpdate) error {
	return uc.contestRepo.UpdateParticipant(ctx, id, participantUpdate)
}
