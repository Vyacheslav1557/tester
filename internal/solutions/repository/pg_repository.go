package repository

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/pkg"
	"github.com/jmoiron/sqlx"
)

type PgRepository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *PgRepository {
	return &PgRepository{
		db: db,
	}
}

const (
	GetSolutionQuery = `
SELECT s.id,

       s.user_id,
       u.username,

       s.solution,

       s.state,
       s.score,
       s.penalty,
       s.time_stat,
       s.memory_stat,
       s.language,

       s.problem_id,
       p.title problem_title,

       cp.position,

       s.contest_id,
       c.title contest_title,

       s.updated_at,
       s.created_at
FROM solutions s
         LEFT JOIN users u ON s.user_id = u.id
         LEFT JOIN problems p ON s.problem_id = p.id
         LEFT JOIN contest_problem cp ON p.id = cp.problem_id AND cp.contest_id = s.contest_id
         LEFT JOIN contests c ON s.contest_id = c.id
WHERE s.id = $1`
)

func (r *PgRepository) GetSolution(ctx context.Context, id int32) (*models.Solution, error) {
	const op = "Repository.GetSolution"

	var solution models.Solution
	err := r.db.GetContext(ctx, &solution, GetSolutionQuery, id)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	return &solution, nil
}

const (
	CreateSolutionQuery = `
INSERT INTO solutions (contest_id, problem_id, user_id, solution, language, penalty) 
VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
)

func (r *PgRepository) CreateSolution(ctx context.Context, creation *models.SolutionCreation) (int32, error) {
	const op = "Repository.CreateSolution"

	rows, err := r.db.QueryxContext(ctx,
		CreateSolutionQuery,
		creation.ContestId,
		creation.ProblemId,
		creation.UserId,
		creation.Solution,
		creation.Language,
		creation.Penalty,
	)
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

const UpdateSolutionQuery = `	
UPDATE solutions
SET state = $1, score = $2, time_stat = $3, memory_stat = $4
WHERE id = $5`

func (r *PgRepository) UpdateSolution(ctx context.Context, id int32, update *models.SolutionUpdate) error {
	const op = "Repository.UpdateSolution"

	_, err := r.db.ExecContext(ctx, UpdateSolutionQuery, update.State, update.Score, update.TimeStat, update.MemoryStat, id)
	if err != nil {
		return pkg.HandlePgErr(err, op)
	}

	return nil
}

func buildListSolutionsQueries(filter models.SolutionsFilter) (sq.SelectBuilder, sq.SelectBuilder) {
	columns := []string{
		"s.id",

		"s.user_id",
		"u.username",

		"s.state",
		"s.score",
		"s.penalty",
		"s.time_stat",
		"s.memory_stat",
		"s.language",

		"s.problem_id",
		"p.title problem_title",

		"cp.position",

		"s.contest_id",
		"c.title contest_title",

		"s.updated_at",
		"s.created_at",
	}

	qb := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select(columns...).
		From("solutions s").
		LeftJoin("users u ON s.user_id = u.id").
		LeftJoin("problems p ON s.problem_id = p.id").
		LeftJoin("contest_problem cp ON p.id = cp.problem_id AND cp.contest_id = s.contest_id").
		LeftJoin("contests c ON s.contest_id = c.id")

	if filter.ContestId != nil {
		qb = qb.Where(sq.Eq{"s.contest_id": *filter.ContestId})
	}
	if filter.UserId != nil {
		qb = qb.Where(sq.Eq{"s.user_id": *filter.UserId})
	}
	if filter.ProblemId != nil {
		qb = qb.Where(sq.Eq{"s.problem_id": *filter.ProblemId})
	}
	if filter.Language != nil {
		qb = qb.Where(sq.Eq{"s.language": *filter.Language})
	}
	if filter.State != nil {
		qb = qb.Where(sq.Eq{"s.state": *filter.State})
	}

	countQb := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("COUNT(*)").FromSelect(qb, "sub")

	if filter.Order != nil && *filter.Order < 0 {
		qb = qb.OrderBy("s.id DESC")
	} else {
		qb = qb.OrderBy("s.id ASC")
	}

	qb = qb.Limit(uint64(filter.PageSize)).Offset(uint64(filter.Offset()))

	return qb, countQb
}

func (r *PgRepository) ListSolutions(ctx context.Context, filter models.SolutionsFilter) (*models.SolutionsList, error) {
	const op = "ContestRepository.ListSolutions"

	baseQb, countQb := buildListSolutionsQueries(filter)

	query, args, err := countQb.ToSql()
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	var totalCount int32
	err = r.db.GetContext(ctx, &totalCount, query, args...)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	query, args, err = baseQb.ToSql()
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}
	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}
	defer rows.Close()

	solutions := make([]*models.SolutionsListItem, 0)
	for rows.Next() {
		var solution models.SolutionsListItem
		err = rows.StructScan(&solution)
		if err != nil {
			return nil, pkg.HandlePgErr(err, op)
		}
		solutions = append(solutions, &solution)
	}

	if err = rows.Err(); err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	return &models.SolutionsList{
		Solutions: solutions,
		Pagination: models.Pagination{
			Total: models.Total(totalCount, filter.PageSize),
			Page:  filter.Page,
		},
	}, nil
}
