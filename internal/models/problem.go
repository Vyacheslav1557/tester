package models

import "time"

type Problem struct {
	Id          int32  `db:"id"`
	Title       string `db:"title"`
	TimeLimit   int32  `db:"time_limit"`
	MemoryLimit int32  `db:"memory_limit"`

	Legend       string `db:"legend"`
	InputFormat  string `db:"input_format"`
	OutputFormat string `db:"output_format"`
	Notes        string `db:"notes"`
	Scoring      string `db:"scoring"`

	LegendHtml       string `db:"legend_html"`
	InputFormatHtml  string `db:"input_format_html"`
	OutputFormatHtml string `db:"output_format_html"`
	NotesHtml        string `db:"notes_html"`
	ScoringHtml      string `db:"scoring_html"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type ProblemsListItem struct {
	Id          int32     `db:"id"`
	Title       string    `db:"title"`
	MemoryLimit int32     `db:"memory_limit"`
	TimeLimit   int32     `db:"time_limit"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	SolvedCount int32     `db:"solved_count"`
}

type ProblemsList struct {
	Problems   []*ProblemsListItem `json:"problems"`
	Pagination Pagination          `json:"pagination"`
}

type ProblemsFilter struct {
	Page     int32
	PageSize int32
}

func (f ProblemsFilter) Offset() int32 {
	return (f.Page - 1) * f.PageSize
}

type ProblemUpdate struct {
	Title       *string `db:"title"`
	MemoryLimit *int32  `db:"memory_limit"`
	TimeLimit   *int32  `db:"time_limit"`

	Legend       *string `db:"legend"`
	InputFormat  *string `db:"input_format"`
	OutputFormat *string `db:"output_format"`
	Notes        *string `db:"notes"`
	Scoring      *string `db:"scoring"`

	LegendHtml       *string `db:"legend_html"`
	InputFormatHtml  *string `db:"input_format_html"`
	OutputFormatHtml *string `db:"output_format_html"`
	NotesHtml        *string `db:"notes_html"`
	ScoringHtml      *string `db:"scoring_html"`
}

type ProblemStatement struct {
	Legend       string `db:"legend"`
	InputFormat  string `db:"input_format"`
	OutputFormat string `db:"output_format"`
	Notes        string `db:"notes"`
	Scoring      string `db:"scoring"`
}

type Html5ProblemStatement struct {
	LegendHtml       string `db:"legend_html"`
	InputFormatHtml  string `db:"input_format_html"`
	OutputFormatHtml string `db:"output_format_html"`
	NotesHtml        string `db:"notes_html"`
	ScoringHtml      string `db:"scoring_html"`
}
