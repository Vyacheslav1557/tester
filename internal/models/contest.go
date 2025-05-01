package models

import "time"

type Contest struct {
	Id        int32     `db:"id"`
	Title     string    `db:"title"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type ContestsListItem struct {
	Id        int32     `db:"id"`
	Title     string    `db:"title"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type ContestsList struct {
	Contests   []*ContestsListItem
	Pagination Pagination
}

type ContestsFilter struct {
	Page     int32
	PageSize int32
	UserId   *int32
	Order    *int32
}

func (f ContestsFilter) Offset() int32 {
	return (f.Page - 1) * f.PageSize
}

type ContestUpdate struct {
	Title *string `json:"title"`
}

type Monitor struct {
	Participants []*ParticipantsStat
	Summary      []*ProblemStatSummary
}

type ParticipantsStat struct {
	Id             int32                `db:"id"`
	Name           string               `db:"name"`
	SolvedInTotal  int32                `db:"solved_in_total"`
	PenaltyInTotal int32                `db:"penalty_in_total"`
	Solutions      []*SolutionsListItem `db:"solutions"`
}

type ProblemStatSummary struct {
	Id       int32 `db:"task_id"`
	Position int32 `db:"position"`
	Success  int32 `db:"success"`
	Total    int32 `db:"total"`
}

type Task struct {
	Id          int32  `db:"id"`
	Position    int32  `db:"position"`
	Title       string `db:"title"`
	TimeLimit   int32  `db:"time_limit"`
	MemoryLimit int32  `db:"memory_limit"`

	ProblemId int32 `db:"problem_id"`
	ContestId int32 `db:"contest_id"`

	LegendHtml       string `db:"legend_html"`
	InputFormatHtml  string `db:"input_format_html"`
	OutputFormatHtml string `db:"output_format_html"`
	NotesHtml        string `db:"notes_html"`
	ScoringHtml      string `db:"scoring_html"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type TasksListItem struct {
	Id          int32     `db:"id"`
	ProblemId   int32     `db:"problem_id"`
	ContestId   int32     `db:"contest_id"`
	Position    int32     `db:"position"`
	Title       string    `db:"title"`
	MemoryLimit int32     `db:"memory_limit"`
	TimeLimit   int32     `db:"time_limit"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type Participant struct {
	Id        int32     `db:"id"`
	UserId    int32     `db:"user_id"`
	ContestId int32     `db:"contest_id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type ParticipantsListItem struct {
	Id        int32     `db:"id"`
	UserId    int32     `db:"user_id"`
	ContestId int32     `db:"contest_id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type ParticipantsList struct {
	Participants []*ParticipantsListItem
	Pagination   Pagination
}

type ParticipantsFilter struct {
	Page     int32
	PageSize int32

	ContestId int32
}

func (f ParticipantsFilter) Offset() int32 {
	return (f.Page - 1) * f.PageSize
}

type ParticipantUpdate struct {
	Name *string `json:"name"`
}
