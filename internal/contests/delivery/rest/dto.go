package rest

import (
	testerv1 "github.com/Vyacheslav1557/tester/contracts/tester/v1"
	"github.com/Vyacheslav1557/tester/internal/models"
)

func GetContestResponseDTO(contest *models.Contest,
	tasks []*models.TasksListItem,
	solutions []*models.SolutionsListItem) *testerv1.GetContestResponse {

	m := make(map[int32]*models.SolutionsListItem)

	for i := 0; i < len(solutions); i++ {
		m[solutions[i].TaskPosition] = solutions[i]
	}

	resp := testerv1.GetContestResponse{
		Contest: ContestDTO(*contest),
		Tasks: make([]struct {
			Solution testerv1.SolutionsListItem `json:"solution"`
			Task     testerv1.TasksListItem     `json:"task"`
		}, len(tasks)),
	}

	for i, task := range tasks {
		solution := testerv1.SolutionsListItem{}
		if sol, ok := m[task.Position]; ok {
			solution = SolutionsListItemDTO(*sol)
		}
		resp.Tasks[i] = struct {
			Solution testerv1.SolutionsListItem `json:"solution"`
			Task     testerv1.TasksListItem     `json:"task"`
		}{
			Solution: solution,
			Task:     TasksListItemDTO(*task),
		}
	}

	return &resp
}

func ListContestsResponseDTO(contestsList *models.ContestsList) *testerv1.ListContestsResponse {
	resp := testerv1.ListContestsResponse{
		Contests:   make([]testerv1.ContestsListItem, len(contestsList.Contests)),
		Pagination: PaginationDTO(contestsList.Pagination),
	}

	for i, contest := range contestsList.Contests {
		resp.Contests[i] = ContestsListItemDTO(*contest)
	}

	return &resp
}

func ListSolutionsResponseDTO(solutionsList *models.SolutionsList) *testerv1.ListSolutionsResponse {
	resp := testerv1.ListSolutionsResponse{
		Solutions:  make([]testerv1.SolutionsListItem, len(solutionsList.Solutions)),
		Pagination: PaginationDTO(solutionsList.Pagination),
	}

	for i, solution := range solutionsList.Solutions {
		resp.Solutions[i] = SolutionsListItemDTO(*solution)
	}

	return &resp
}

func GetTaskResponseDTO(contest *models.Contest, tasks []*models.TasksListItem, task *models.Task) *testerv1.GetTaskResponse {
	resp := testerv1.GetTaskResponse{
		Contest: ContestDTO(*contest),
		Tasks:   make([]testerv1.TasksListItem, len(tasks)),
		Task:    *TaskDTO(task),
	}

	for i, t := range tasks {
		resp.Tasks[i] = TasksListItemDTO(*t)
	}

	return &resp
}

func PaginationDTO(p models.Pagination) testerv1.Pagination {
	return testerv1.Pagination{
		Page:  p.Page,
		Total: p.Total,
	}
}

func ContestDTO(c models.Contest) testerv1.Contest {
	return testerv1.Contest{
		Id:        c.Id,
		Title:     c.Title,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func ContestsListItemDTO(c models.ContestsListItem) testerv1.ContestsListItem {
	return testerv1.ContestsListItem{
		Id:        c.Id,
		Title:     c.Title,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func TasksListItemDTO(t models.TasksListItem) testerv1.TasksListItem {
	return testerv1.TasksListItem{
		Id:          t.Id,
		Position:    t.Position,
		Title:       t.Title,
		MemoryLimit: t.MemoryLimit,
		ProblemId:   t.ProblemId,
		TimeLimit:   t.TimeLimit,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func TaskDTO(t *models.Task) *testerv1.Task {
	return &testerv1.Task{
		Id:          t.Id,
		Title:       t.Title,
		MemoryLimit: t.MemoryLimit,
		TimeLimit:   t.TimeLimit,

		InputFormatHtml:  t.InputFormatHtml,
		LegendHtml:       t.LegendHtml,
		NotesHtml:        t.NotesHtml,
		OutputFormatHtml: t.OutputFormatHtml,
		Position:         t.Position,
		ScoringHtml:      t.ScoringHtml,

		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func ParticipantsListItemDTO(p models.ParticipantsListItem) testerv1.ParticipantsListItem {
	return testerv1.ParticipantsListItem{
		Id:        p.Id,
		UserId:    p.UserId,
		Name:      p.Name,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func SolutionsListItemDTO(s models.SolutionsListItem) testerv1.SolutionsListItem {
	return testerv1.SolutionsListItem{
		Id: s.Id,

		ParticipantId:   s.ParticipantId,
		ParticipantName: s.ParticipantName,

		State:      s.State,
		Score:      s.Score,
		Penalty:    s.Penalty,
		TimeStat:   s.TimeStat,
		MemoryStat: s.MemoryStat,
		Language:   s.Language,

		TaskId:       s.TaskId,
		TaskPosition: s.TaskPosition,
		TaskTitle:    s.TaskTitle,

		ContestId:    s.ContestId,
		ContestTitle: s.ContestTitle,

		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

func SolutionDTO(s models.Solution) testerv1.Solution {
	return testerv1.Solution{
		Id: s.Id,

		ParticipantId:   s.ParticipantId,
		ParticipantName: s.ParticipantName,

		Solution: s.Solution,

		State:      s.State,
		Score:      s.Score,
		Penalty:    s.Penalty,
		TimeStat:   s.TimeStat,
		MemoryStat: s.MemoryStat,
		Language:   s.Language,

		TaskId:       s.TaskId,
		TaskPosition: s.TaskPosition,
		TaskTitle:    s.TaskTitle,

		ContestId:    s.ContestId,
		ContestTitle: s.ContestTitle,

		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}
