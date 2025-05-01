package contests

import (
	"context"
	"github.com/Vyacheslav1557/tester/internal/models"
)

type Repository interface {
	CreateContest(ctx context.Context, title string) (int32, error)
	GetContest(ctx context.Context, id int32) (*models.Contest, error)
	DeleteContest(ctx context.Context, id int32) error
	UpdateContest(ctx context.Context, id int32, contestUpdate models.ContestUpdate) error
	ListContests(ctx context.Context, filter models.ContestsFilter) (*models.ContestsList, error)

	CreateTask(ctx context.Context, contestId int32, taskId int32) (int32, error)
	GetTask(ctx context.Context, id int32) (*models.Task, error)
	DeleteTask(ctx context.Context, taskId int32) error
	GetTasks(ctx context.Context, contestId int32) ([]*models.TasksListItem, error)

	GetParticipantId(ctx context.Context, contestId int32, userId int32) (int32, error)
	GetParticipantId2(ctx context.Context, taskId int32, userId int32) (int32, error)
	GetParticipantId3(ctx context.Context, solutionId int32) (int32, error)
	CreateParticipant(ctx context.Context, contestId int32, userId int32) (int32, error)
	DeleteParticipant(ctx context.Context, participantId int32) error
	UpdateParticipant(ctx context.Context, id int32, participantUpdate models.ParticipantUpdate) error
	ListParticipants(ctx context.Context, filter models.ParticipantsFilter) (*models.ParticipantsList, error)

	GetSolution(ctx context.Context, id int32) (*models.Solution, error)
	CreateSolution(ctx context.Context, creation *models.SolutionCreation) (int32, error)
	ListSolutions(ctx context.Context, filter models.SolutionsFilter) (*models.SolutionsList, error)
	GetBestSolutions(ctx context.Context, contestId int32, participantId int32) ([]*models.SolutionsListItem, error)

	GetMonitor(ctx context.Context, id int32, penalty int32) (*models.Monitor, error)
}
