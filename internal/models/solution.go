package models

import "time"

type Solution struct {
	Id int32 `db:"id"`

	ParticipantId   int32  `db:"participant_id"`
	ParticipantName string `db:"participant_name"`

	Solution string `db:"solution"`

	State      int32 `db:"state"`
	Score      int32 `db:"score"`
	Penalty    int32 `db:"penalty"`
	TimeStat   int32 `db:"time_stat"`
	MemoryStat int32 `db:"memory_stat"`
	Language   int32 `db:"language"`

	TaskId       int32  `db:"task_id"`
	TaskPosition int32  `db:"task_position"`
	TaskTitle    string `db:"task_title"`

	ContestId    int32  `db:"contest_id"`
	ContestTitle string `db:"contest_title"`

	UpdatedAt time.Time `db:"updated_at"`
	CreatedAt time.Time `db:"created_at"`
}

type SolutionCreation struct {
	Solution      string
	TaskId        int32
	UserId        int32
	ParticipantId int32
	Language      int32
	Penalty       int32
}

type SolutionsListItem struct {
	Id int32 `db:"id"`

	ParticipantId   int32  `db:"participant_id"`
	ParticipantName string `db:"participant_name"`

	State      int32 `db:"state"`
	Score      int32 `db:"score"`
	Penalty    int32 `db:"penalty"`
	TimeStat   int32 `db:"time_stat"`
	MemoryStat int32 `db:"memory_stat"`
	Language   int32 `db:"language"`

	TaskId       int32  `db:"task_id"`
	TaskPosition int32  `db:"task_position"`
	TaskTitle    string `db:"task_title"`

	ContestId    int32  `db:"contest_id"`
	ContestTitle string `db:"contest_title"`

	UpdatedAt time.Time `db:"updated_at"`
	CreatedAt time.Time `db:"created_at"`
}

type SolutionsList struct {
	Solutions  []*SolutionsListItem
	Pagination Pagination
}

type SolutionsFilter struct {
	Page          int32
	PageSize      int32
	ContestId     *int32
	ParticipantId *int32
	TaskId        *int32
	Language      *int32
	State         *int32
	Order         *int32
}

func (f SolutionsFilter) Offset() int32 {
	return (f.Page - 1) * f.PageSize
}

//type Result int32
//
//const (
//	NotTested               Result = 1 // change only with schema change
//	Accepted                Result = 2
//	WrongAnswer             Result = 3
//	PresentationError       Result = 4
//	CompilationError        Result = 5
//	MemoryLimitExceeded     Result = 6
//	TimeLimitExceeded       Result = 7
//	RuntimeError            Result = 8
//	SystemFailDuringTesting Result = 9
//	Testing                 Result = 10
//)
//
//var ErrBadResult = errors.New("bad result")
//
//func (result Result) Valid() error {
//	switch result {
//	case NotTested, Accepted, TimeLimitExceeded, MemoryLimitExceeded, CompilationError, SystemFailDuringTesting:
//		return nil
//	}
//	return ErrBadResult
//}
//
//type Language struct {
//	Name       string
//	CompileCmd []string //source: src;result:executable
//	RunCmd     []string //source: executable
//}
//
//var Languages = []Language{
//	{Name: "gcc std=c90",
//		CompileCmd: []string{"gcc", "src", "-std=c90", "-o", "executable"},
//		RunCmd:     []string{"executable"}},
//}
