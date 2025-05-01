package repository

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/pkg"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		db: db,
	}
}

const CreateContestQuery = "INSERT INTO contests (title) VALUES ($1) RETURNING id"

func (r *Repository) CreateContest(ctx context.Context, title string) (int32, error) {
	const op = "Repository.CreateContest"

	rows, err := r.db.QueryxContext(ctx, CreateContestQuery, title)
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

const GetContestQuery = "SELECT * from contests WHERE id=$1 LIMIT 1"

func (r *Repository) GetContest(ctx context.Context, id int32) (*models.Contest, error) {
	const op = "Repository.GetContest"

	var contest models.Contest
	err := r.db.GetContext(ctx, &contest, GetContestQuery, id)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}
	return &contest, nil
}

const (
	UpdateContestQuery = "UPDATE contests SET title = COALESCE($1, title) WHERE id = $2"
)

func (r *Repository) UpdateContest(ctx context.Context, id int32, contestUpdate models.ContestUpdate) error {
	const op = "Repository.UpdateContest"

	_, err := r.db.ExecContext(ctx, UpdateContestQuery, contestUpdate.Title, id)
	if err != nil {
		return pkg.HandlePgErr(err, op)
	}

	return nil
}

const DeleteContestQuery = "DELETE FROM contests WHERE id=$1"

func (r *Repository) DeleteContest(ctx context.Context, id int32) error {
	const op = "Repository.DeleteContest"

	_, err := r.db.ExecContext(ctx, DeleteContestQuery, id)
	if err != nil {
		return pkg.HandlePgErr(err, op)
	}

	return nil
}

func buildListContestsQueries(filter models.ContestsFilter) (sq.SelectBuilder, sq.SelectBuilder) {
	columns := []string{
		"c.id",
		"c.title",
		"c.created_at",
		"c.updated_at",
	}

	qb := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select(columns...).From("contests c")

	if filter.UserId != nil {
		qb = qb.Join("participants p ON c.id = p.contest_id")
		qb = qb.Where(sq.Eq{"p.user_id": *filter.UserId})
	}

	countQb := sq.Select("COUNT(*)").FromSelect(qb, "sub")

	if filter.Order != nil && *filter.Order < 0 {
		qb = qb.OrderBy("c.created_at DESC")
	} else {
		qb = qb.OrderBy("c.created_at ASC")
	}

	qb = qb.Limit(uint64(filter.PageSize)).Offset(uint64(filter.Offset()))

	return qb, countQb
}

func (r *Repository) ListContests(ctx context.Context, filter models.ContestsFilter) (*models.ContestsList, error) {
	const op = "Repository.ListContests"

	baseQb, countQb := buildListContestsQueries(filter)

	query, args, err := baseQb.ToSql()
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	var contests []*models.ContestsListItem
	err = r.db.SelectContext(ctx, &contests, query, args...)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	query, args, err = countQb.ToSql()
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	var count int32
	err = r.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	return &models.ContestsList{
		Contests: contests,
		Pagination: models.Pagination{
			Total: models.Total(count, filter.PageSize),
			Page:  filter.Page,
		},
	}, nil
}
