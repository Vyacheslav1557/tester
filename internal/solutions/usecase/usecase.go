package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/internal/problems"
	"github.com/Vyacheslav1557/tester/internal/solutions"
	"github.com/Vyacheslav1557/tester/pkg/tester"
	"time"
)

type Publisher interface {
	Publish(subject string, data []byte) error
}

type Tester interface {
	Test(ctx context.Context, packet tester.Packet, s tester.Solution) <-chan tester.TestingMessage
}

type UseCase struct {
	solutionsRepo solutions.Repository
	problemsUC    problems.UseCase
	pub           Publisher
	tester        Tester
}

func NewUseCase(
	solutionsRepo solutions.Repository,
	problemsUC problems.UseCase,
	pub Publisher,
	tester Tester,
) *UseCase {
	return &UseCase{
		solutionsRepo: solutionsRepo,
		problemsUC:    problemsUC,
		pub:           pub,
		tester:        tester,
	}
}

func (uc *UseCase) GetSolution(ctx context.Context, id int32) (*models.Solution, error) {
	return uc.solutionsRepo.GetSolution(ctx, id)
}

func (uc *UseCase) CreateSolution(ctx context.Context, creation *models.SolutionCreation) (int32, error) {
	id, err := uc.solutionsRepo.CreateSolution(ctx, creation)
	if err != nil {
		return 0, err
	}

	problem, err := uc.problemsUC.GetProblemById(ctx, creation.ProblemId)
	if err != nil {
		return 0, err
	}

	// if there are no tests, just accept the solution
	if problem.Meta.Count == 0 {
		err := uc.solutionsRepo.UpdateSolution(ctx, id, &models.SolutionUpdate{
			State:      models.Accepted,
			Score:      100,
			TimeStat:   0,
			MemoryStat: 0,
		})
		if err != nil {
			return 0, err
		}

		return 0, nil
	}

	zipPath, err := uc.problemsUC.DownloadTestsArchive(ctx, problem.Id)
	if err != nil {
		return 0, err
	}

	packet := Packet{
		contestId:   creation.ContestId,
		problemId:   problem.Id,
		updatedAt:   problem.UpdatedAt.Unix(),
		zipPath:     zipPath,
		timeLimit:   int64(problem.TimeLimit),
		memoryLimit: int64(problem.MemoryLimit),
		meta:        &problem.Meta,
	}

	solution := &Solution{
		solution: []byte(creation.Solution),
		language: creation.Language,
		id:       id,
	}

	sol, err := uc.solutionsRepo.GetSolution(ctx, id)
	if err != nil {
		return 0, err
	}

	go uc.test(ctx, packet, solution, sol)

	return id, err
}

func (uc *UseCase) UpdateSolution(ctx context.Context, id int32, update *models.SolutionUpdate) error {
	return uc.solutionsRepo.UpdateSolution(ctx, id, update)
}

func (uc *UseCase) ListSolutions(ctx context.Context, filter models.SolutionsFilter) (*models.SolutionsList, error) {
	return uc.solutionsRepo.ListSolutions(ctx, filter)
}

func (uc *UseCase) test(ctx context.Context, packet tester.Packet, s tester.Solution, sol *models.Solution) {
	ch := uc.tester.Test(ctx, packet, s)

	sli := SolutionsListItem{
		Id: sol.Id,

		UserId:   sol.UserId,
		Username: sol.Username,

		State:      sol.State,
		Score:      sol.Score,
		Penalty:    sol.Penalty,
		TimeStat:   sol.TimeStat,
		MemoryStat: sol.MemoryStat,
		Language:   sol.Language,

		ProblemId:    sol.ProblemId,
		ProblemTitle: sol.ProblemTitle,

		Position: sol.Position,

		ContestId:    sol.ContestId,
		ContestTitle: sol.ContestTitle,

		UpdatedAt: sol.UpdatedAt,
		CreatedAt: sol.CreatedAt,
	}

	uc.publish(packet.ContestId(), &Message{
		MessageType: MessageTypeCreate,
		Solution:    sli,
	})

	solutionUpdate := models.SolutionUpdate{
		State:      models.Saved,
		Score:      0,
		TimeStat:   0,
		MemoryStat: 0,
	}

	testsPassed := 0
	testsPassedExpected := packet.Meta().Count

	for msg := range ch {
		if msg.Details != "" {
			uc.publish(packet.ContestId(), &Message{
				MessageType: MessageTypeUpdate,
				Solution:    sli,
				Message:     &msg.Details,
			})
		}

		if msg.Err != nil {
			var stErr *tester.StateErr
			if errors.As(msg.Err, &stErr) {
				solutionUpdate.State = stErr.State
				break
			}

			fmt.Println("something really bad happened here")
			break
		}

		if msg.Metrics != nil {
			testsPassed++

			// doing this way we get the max over all tests
			solutionUpdate.MemoryStat = max(
				solutionUpdate.MemoryStat,
				int32(msg.Metrics.MaximumResidentSetSize),
			)
			solutionUpdate.TimeStat = max(
				solutionUpdate.TimeStat,
				int32(msg.Metrics.ElapsedTime.Milliseconds()),
			)
		}
	}

	if testsPassed != testsPassedExpected && solutionUpdate.State == models.Saved {
		fmt.Println("something bad had happened")
		return
	}

	if testsPassed == testsPassedExpected {
		solutionUpdate.Score = 100
		solutionUpdate.State = models.Accepted
	}

	if err := uc.solutionsRepo.UpdateSolution(ctx, s.Id(), &solutionUpdate); err != nil {
		fmt.Println(err)
		return
	}

	sli.State = solutionUpdate.State
	sli.Score = solutionUpdate.Score
	sli.TimeStat = solutionUpdate.TimeStat
	sli.MemoryStat = solutionUpdate.MemoryStat

	uc.publish(packet.ContestId(),
		&Message{
			MessageType: MessageTypeUpdate,
			Solution:    sli,
		})
}

func (uc *UseCase) publish(contestId int32, msg *Message) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return uc.pub.Publish(fmt.Sprintf("contest-%d-solutions", contestId), b)
}

type Packet struct {
	contestId   int32
	problemId   int32
	updatedAt   int64
	zipPath     string
	timeLimit   int64
	memoryLimit int64
	meta        *models.Meta
}

func (p Packet) ContestId() int32 {
	return p.contestId
}

func (p Packet) UniquePacketName() string {
	return fmt.Sprintf("%d_%d", p.problemId, p.updatedAt)
}

func (p Packet) ZipPath() string {
	return p.zipPath
}

func (p Packet) TL() int64 {
	return p.timeLimit
}

func (p Packet) ML() int64 {
	return p.memoryLimit
}

func (p Packet) Meta() *models.Meta {
	return p.meta
}

type Solution struct {
	solution []byte
	language models.LanguageName
	id       int32
}

func (s *Solution) Solution() []byte {
	return s.solution
}

func (s *Solution) Lang() models.LanguageName {
	return s.language
}

func (s *Solution) Id() int32 {
	return s.id
}

const (
	MessageTypeCreate = "CREATE"
	MessageTypeUpdate = "UPDATE"
	MessageTypeDelete = "DELETE"
)

type SolutionsListItem struct {
	Id int32 `json:"id"`

	UserId   int32  `json:"user_id"`
	Username string `json:"username"`

	State      models.State        `json:"state"`
	Score      int32               `json:"score"`
	Penalty    int32               `json:"penalty"`
	TimeStat   int32               `json:"time_stat"`
	MemoryStat int32               `json:"memory_stat"`
	Language   models.LanguageName `json:"language"`

	ProblemId    int32  `json:"problem_id"`
	ProblemTitle string `json:"problem_title"`

	Position int32 `json:"position"`

	ContestId    int32  `json:"contest_id"`
	ContestTitle string `json:"contest_title"`

	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

type Message struct {
	MessageType string            `json:"message_type"`
	Message     *string           `json:"message,omitempty"`
	Solution    SolutionsListItem `json:"solution"`
}
