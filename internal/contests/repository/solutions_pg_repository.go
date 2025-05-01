package repository

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/pkg"
)

const (
	GetSolutionQuery = "SELECT * FROM solutions WHERE id = $1"
)

func (r *Repository) GetSolution(ctx context.Context, id int32) (*models.Solution, error) {
	const op = "Repository.GetSolution"

	var solution models.Solution
	err := r.db.GetContext(ctx, &solution, GetSolutionQuery, id)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	return &solution, nil
}

const (
	CreateSolutionQuery = `INSERT INTO solutions (task_id, participant_id, language, penalty, solution)
VALUES ($1, $2, $3, $4, $5)
RETURNING id`
)

func (r *Repository) CreateSolution(ctx context.Context, creation *models.SolutionCreation) (int32, error) {
	const op = "Repository.CreateSolution"

	rows, err := r.db.QueryxContext(ctx,
		CreateSolutionQuery,
		creation.TaskId,
		creation.ParticipantId,
		creation.Language,
		creation.Penalty,
		creation.Solution,
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

func buildListSolutionsQueries(filter models.SolutionsFilter) (sq.SelectBuilder, sq.SelectBuilder) {
	columns := []string{
		"s.id",
		"s.participant_id",
		"p2.name AS participant_name",
		"s.state",
		"s.score",
		"s.penalty",
		"s.time_stat",
		"s.memory_stat",
		"s.language",
		"s.task_id",
		"t.position AS task_position",
		"p.title AS task_title",
		"t.contest_id",
		"c.title",
		"s.updated_at",
		"s.created_at",
	}

	qb := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select(columns...).
		From("solutions s").
		LeftJoin("tasks t ON s.task_id = t.id").
		LeftJoin("problems p ON t.problem_id = p.id").
		LeftJoin("contests c ON t.contest_id = c.id").
		LeftJoin("participants p2 ON s.participant_id = p2.id")

	if filter.ContestId != nil {
		qb = qb.Where(sq.Eq{"s.contest_id": *filter.ContestId})
	}
	if filter.ParticipantId != nil {
		qb = qb.Where(sq.Eq{"s.participant_id": *filter.ParticipantId})
	}
	if filter.TaskId != nil {
		qb = qb.Where(sq.Eq{"s.task_id": *filter.TaskId})
	}
	if filter.Language != nil {
		qb = qb.Where(sq.Eq{"s.language": *filter.Language})
	}
	if filter.State != nil {
		qb = qb.Where(sq.Eq{"s.state": *filter.State})
	}

	countQb := sq.Select("COUNT(*)").FromSelect(qb, "sub")

	if filter.Order != nil && *filter.Order < 0 {
		qb = qb.OrderBy("s.id DESC")
	} else {
		qb = qb.OrderBy("s.id ASC")
	}

	qb = qb.Limit(uint64(filter.PageSize)).Offset(uint64(filter.Offset()))

	return qb, countQb
}

func (r *Repository) ListSolutions(ctx context.Context, filter models.SolutionsFilter) (*models.SolutionsList, error) {
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

const (
	// state=5 - AC
	GetBestSolutions = `
		WITH contest_tasks AS (
    SELECT t.id AS task_id,
           t.position AS task_position,
           t.contest_id,
           t.problem_id,
           t.created_at,
           t.updated_at,
           p.title AS task_title,
           c.title AS contest_title
    FROM tasks t
             LEFT JOIN problems p ON p.id = t.problem_id
             LEFT JOIN  contests c ON c.id = t.contest_id
    WHERE t.contest_id = ?
),
     best_solutions AS (
         SELECT DISTINCT ON (s.task_id)
             *
         FROM solutions s
         WHERE s.participant_id = ?
         ORDER BY s.task_id, s.score DESC, s.created_at DESC
     )
SELECT
    s.id,
    s.participant_id,
    p.name AS participant_name,
    s.state,
    s.score,
    s.penalty,
    s.time_stat,
    s.memory_stat,
    s.language,
    ct.task_id,
    ct.task_position,
    ct.task_title,
    ct.contest_id,
    ct.contest_title,
    s.updated_at,
    s.created_at
FROM contest_tasks ct
         LEFT JOIN best_solutions s ON s.task_id = ct.task_id
         LEFT JOIN participants p ON p.id = s.participant_id WHERE s.id IS NOT NULL
ORDER BY ct.task_position
`
)

func (r *Repository) GetBestSolutions(ctx context.Context, contestId int32, participantId int32) ([]*models.SolutionsListItem, error) {
	const op = "Repository.GetBestSolutions"
	var solutions []*models.SolutionsListItem
	query := r.db.Rebind(GetBestSolutions)
	err := r.db.SelectContext(ctx, &solutions, query, contestId, participantId)

	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	return solutions, nil
}
