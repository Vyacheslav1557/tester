package usecase

import (
	"context"
	"github.com/Vyacheslav1557/tester/internal/models"
)

func (uc *ContestUseCase) CreateTask(ctx context.Context, contestId int32, taskId int32) (id int32, err error) {
	return uc.contestRepo.CreateTask(ctx, contestId, taskId)
}

func (uc *ContestUseCase) GetTask(ctx context.Context, id int32) (*models.Task, error) {
	return uc.contestRepo.GetTask(ctx, id)
}

func (uc *ContestUseCase) GetTasks(ctx context.Context, contestId int32) ([]*models.TasksListItem, error) {
	return uc.contestRepo.GetTasks(ctx, contestId)
}

func (uc *ContestUseCase) DeleteTask(ctx context.Context, taskId int32) error {
	return uc.contestRepo.DeleteTask(ctx, taskId)
}
