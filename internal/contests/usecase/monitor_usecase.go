package usecase

import (
	"context"
	"github.com/Vyacheslav1557/tester/internal/models"
)

func (uc *ContestUseCase) GetMonitor(ctx context.Context, contestId int32) (*models.Monitor, error) {
	return uc.contestRepo.GetMonitor(ctx, contestId, 20)
}
