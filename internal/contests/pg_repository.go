package contests

import (
	"context"
	"github.com/Vyacheslav1557/tester/internal/models"
)

type Repository interface {
	CreateContest(ctx context.Context, title string) (int32, error)
	GetContest(ctx context.Context, id int32) (*models.Contest, error)
	ListContests(ctx context.Context, filter models.ContestsFilter) (*models.ContestsList, error)
	UpdateContest(ctx context.Context, id int32, contestUpdate models.ContestUpdate) error
	DeleteContest(ctx context.Context, id int32) error

	CreateContestProblem(ctx context.Context, contestId, problemId int32) error
	GetContestProblem(ctx context.Context, contestId int32, problemId int32) (*models.ContestProblem, error)
	GetContestProblems(ctx context.Context, contestId int32) ([]*models.ContestProblemsListItem, error)
	DeleteContestProblem(ctx context.Context, contestId, problemId int32) error

	CreateParticipant(ctx context.Context, contestId, userId int32) error
	IsParticipant(ctx context.Context, contestId int32, userId int32) (bool, error)
	DeleteParticipant(ctx context.Context, contestId, userId int32) error
	ListParticipants(ctx context.Context, filter models.ParticipantsFilter) (*models.UsersList, error)

	GetMonitor(ctx context.Context, contestId int32) (*models.Monitor, error)
}
