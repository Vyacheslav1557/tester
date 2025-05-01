package repository

import (
	"context"
	"github.com/Vyacheslav1557/tester/internal/problems"
	"github.com/Vyacheslav1557/tester/pkg"

	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	_db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		_db: db,
	}
}

func (r *Repository) BeginTx(ctx context.Context) (problems.Tx, error) {
	tx, err := r._db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (r *Repository) DB() problems.Querier {
	return r._db
}

const CreateProblemQuery = "INSERT INTO problems (title) VALUES ($1) RETURNING id"

func (r *Repository) CreateProblem(ctx context.Context, q problems.Querier, title string) (int32, error) {
	const op = "Repository.CreateProblem"

	rows, err := q.QueryxContext(ctx, CreateProblemQuery, title)
	if err != nil {
		return 0, pkg.HandlePgErr(err, op)
	}

	defer rows.Close()
	var id int32
	rows.Next()
	err = rows.Scan(&id)
	if err != nil {
		return 0, pkg.HandlePgErr(err, op)
	}

	return id, nil
}

const GetProblemByIdQuery = "SELECT * from problems WHERE id=$1 LIMIT 1"

func (r *Repository) GetProblemById(ctx context.Context, q problems.Querier, id int32) (*models.Problem, error) {
	const op = "Repository.ReadProblemById"

	var problem models.Problem
	err := q.GetContext(ctx, &problem, GetProblemByIdQuery, id)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	return &problem, nil
}

const DeleteProblemQuery = "DELETE FROM problems WHERE id=$1"

func (r *Repository) DeleteProblem(ctx context.Context, q problems.Querier, id int32) error {
	const op = "Repository.DeleteProblem"

	_, err := q.ExecContext(ctx, DeleteProblemQuery, id)
	if err != nil {
		return pkg.HandlePgErr(err, op)
	}

	return nil
}

const (
	ListProblemsQuery = `SELECT p.id,
       p.title,
       p.memory_limit,
       p.time_limit,
       p.created_at,
       p.updated_at,
       COALESCE(solved_count, 0) AS solved_count
FROM problems p
         LEFT JOIN (SELECT t.problem_id,
                           COUNT(DISTINCT s.participant_id) AS solved_count
                    FROM solutions s
                             JOIN tasks t ON s.task_id = t.id
                    WHERE s.state = 5
                    GROUP BY t.problem_id) sol ON p.id = sol.problem_id
LIMIT $1 OFFSET $2`
	CountProblemsQuery = "SELECT COUNT(*) FROM problems"
)

func (r *Repository) ListProblems(ctx context.Context, q problems.Querier, filter models.ProblemsFilter) (*models.ProblemsList, error) {
	const op = "ContestRepository.ListProblems"

	var list []*models.ProblemsListItem
	err := q.SelectContext(ctx, &list, ListProblemsQuery, filter.PageSize, filter.Offset())
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	var count int32
	err = q.GetContext(ctx, &count, CountProblemsQuery)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	return &models.ProblemsList{
		Problems: list,
		Pagination: models.Pagination{
			Total: models.Total(count, filter.PageSize),
			Page:  filter.Page,
		},
	}, nil
}

const (
	UpdateProblemQuery = `UPDATE problems
SET title              = COALESCE($2, title),
    time_limit         = COALESCE($3, time_limit),
    memory_limit       = COALESCE($4, memory_limit),

    legend             = COALESCE($5, legend),
    input_format       = COALESCE($6, input_format),
    output_format      = COALESCE($7, output_format),
    notes              = COALESCE($8, notes),
    scoring            = COALESCE($9, scoring),

    legend_html        = COALESCE($10, legend_html),
    input_format_html  = COALESCE($11, input_format_html),
    output_format_html = COALESCE($12, output_format_html),
    notes_html         = COALESCE($13, notes_html),
    scoring_html       = COALESCE($14, scoring_html)

WHERE id=$1`
)

func (r *Repository) UpdateProblem(ctx context.Context, q problems.Querier, id int32, problem *models.ProblemUpdate) error {
	const op = "Repository.UpdateProblem"

	query := q.Rebind(UpdateProblemQuery)
	_, err := q.ExecContext(ctx, query,
		id,

		problem.Title,
		problem.TimeLimit,
		problem.MemoryLimit,

		problem.Legend,
		problem.InputFormat,
		problem.OutputFormat,
		problem.Notes,
		problem.Scoring,

		problem.LegendHtml,
		problem.InputFormatHtml,
		problem.OutputFormatHtml,
		problem.NotesHtml,
		problem.ScoringHtml,
	)
	if err != nil {
		return pkg.HandlePgErr(err, op)
	}

	return nil
}
