package repository

import (
	"context"
	sq "github.com/Masterminds/squirrel"
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

func buildListProblemsQueries(filter models.ProblemsFilter) (sq.SelectBuilder, sq.SelectBuilder) {
	qb := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("id, title, memory_limit, time_limit, created_at, updated_at").From("problems")

	if filter.Title != nil {
		subquery := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
			Select("id, title, memory_limit, time_limit, created_at, updated_at, word_similarity(title, ?) AS similarity").
			From("problems").
			Where("word_similarity(title, ?) > 0", *filter.Title, *filter.Title)

		qb = sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
			Select("id, title, memory_limit, time_limit, created_at, updated_at").
			FromSelect(subquery, "sub").
			OrderBy("similarity DESC")
	}

	countQb := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("COUNT(*)").FromSelect(qb, "sub")

	if filter.Order != nil && *filter.Order < 0 {
		qb = qb.OrderBy("created_at DESC")
	} else {
		qb = qb.OrderBy("created_at ASC")
	}

	qb = qb.Limit(uint64(filter.PageSize)).Offset(uint64(filter.Offset()))

	return qb, countQb
}

func (r *Repository) ListProblems(ctx context.Context, q problems.Querier, filter models.ProblemsFilter) (*models.ProblemsList, error) {
	const op = "ContestRepository.ListProblems"

	ListProblemsQuery, CountProblemsQuery := buildListProblemsQueries(filter)

	query, args, err := ListProblemsQuery.ToSql()
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	list := make([]*models.ProblemsListItem, 0)
	err = q.SelectContext(ctx, &list, query, args...)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	query, args, err = CountProblemsQuery.ToSql()
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	var count int32
	err = q.GetContext(ctx, &count, query, args...)
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
    scoring_html       = COALESCE($14, scoring_html),
    
    meta               = COALESCE($15, meta),
	samples            = COALESCE($16, samples)

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

		problem.Meta,
		problem.Samples,
	)
	if err != nil {
		return pkg.HandlePgErr(err, op)
	}

	return nil
}
