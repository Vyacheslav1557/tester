package repository

import (
	"context"
	"database/sql"
	"errors"
	sq "github.com/Masterminds/squirrel"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/pkg"
	"github.com/jmoiron/sqlx"
	"sort"
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

func (r *Repository) ListContests(ctx context.Context, filter models.ContestsFilter) (*models.ContestsList, error) {
	const op = "Repository.ListContests"

	baseQb, countQb := buildListContestsQueries(filter)

	query, args, err := baseQb.ToSql()
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	contests := make([]*models.Contest, 0)
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

func buildListContestsQueries(filter models.ContestsFilter) (sq.SelectBuilder, sq.SelectBuilder) {
	columns := []string{
		"c.id",
		"c.title",
		"c.created_at",
		"c.updated_at",
	}

	qb := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select(columns...).From("contests c")

	if filter.UserId != nil {
		qb = qb.LeftJoin("contest_user p ON c.id = p.contest_id")
		qb = qb.Where(sq.Eq{"p.user_id": *filter.UserId})
	}

	countQb := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("COUNT(*)").FromSelect(qb, "sub")

	if filter.Order != nil && *filter.Order < 0 {
		qb = qb.OrderBy("c.created_at DESC")
	} else {
		qb = qb.OrderBy("c.created_at ASC")
	}

	qb = qb.Limit(uint64(filter.PageSize)).Offset(uint64(filter.Offset()))

	return qb, countQb
}

const CreateContestProblemQuery = `INSERT INTO contest_problem (problem_id, contest_id, position)
VALUES ($1, $2, COALESCE((SELECT MAX(position) FROM contest_problem WHERE contest_id = $2), 0) + 1)
`

func (r *Repository) CreateContestProblem(ctx context.Context, contestId, problemId int32) error {
	const op = "Repository.CreateContestProblem"

	_, err := r.db.ExecContext(ctx, CreateContestProblemQuery, problemId, contestId)
	if err != nil {
		return pkg.HandlePgErr(err, op)
	}

	return nil
}

const DeleteContestProblemQuery = "DELETE FROM contest_problem WHERE contest_id=$1 AND problem_id=$2"

func (r *Repository) DeleteContestProblem(ctx context.Context, contestId, problemId int32) error {
	const op = "Repository.DeleteContestProblem"

	_, err := r.db.ExecContext(ctx, DeleteContestProblemQuery, contestId, problemId)
	if err != nil {
		return pkg.HandlePgErr(err, op)
	}

	return nil
}

const GetContestProblemQuery = `
SELECT cp.problem_id,
	   p.title,
	   p.time_limit,
	   p.memory_limit,
	   cp.position,
	   p.legend_html,
	   p.input_format_html,
	   p.output_format_html,
	   p.notes_html,
	   p.scoring_html,
	   p.meta,
	   p.samples,
	   p.created_at,
	   p.updated_at
FROM contest_problem cp LEFT JOIN problems p ON cp.problem_id = p.id
WHERE cp.contest_id = $1 AND cp.problem_id = $2
`

func (r *Repository) GetContestProblem(ctx context.Context, contestId, problemId int32) (*models.ContestProblem, error) {
	const op = "Repository.GetContestProblem"

	var contestProblem models.ContestProblem
	err := r.db.GetContext(ctx, &contestProblem, GetContestProblemQuery, contestId, problemId)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	return &contestProblem, nil
}

const GetContestProblemsQuery = `
SELECT cp.problem_id,
	   p.title,
	   p.time_limit,
	   p.memory_limit,
	   cp.position,
	   p.created_at,
	   p.updated_at
FROM contest_problem cp LEFT JOIN problems p ON cp.problem_id = p.id
WHERE cp.contest_id = $1
ORDER BY cp.position
`

func (r *Repository) GetContestProblems(ctx context.Context, contestId int32) ([]*models.ContestProblemsListItem, error) {
	const op = "Repository.GetContestProblems"

	contestProblems := make([]*models.ContestProblemsListItem, 0)
	err := r.db.SelectContext(ctx, &contestProblems, GetContestProblemsQuery, contestId)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	return contestProblems, nil
}

const CreateParticipantQuery = "INSERT INTO contest_user (user_id, contest_id) VALUES ($1, $2)"

func (r *Repository) CreateParticipant(ctx context.Context, contestId int32, userId int32) error {
	const op = "Repository.CreateParticipant"

	_, err := r.db.ExecContext(ctx, CreateParticipantQuery, userId, contestId)
	if err != nil {
		return pkg.HandlePgErr(err, op)
	}

	return nil
}

const DeleteParticipantQuery = "DELETE FROM contest_user WHERE user_id=$1 AND contest_id=$2"

func (r *Repository) DeleteParticipant(ctx context.Context, contestId int32, userId int32) error {
	const op = "Repository.DeleteParticipant"

	_, err := r.db.ExecContext(ctx, DeleteParticipantQuery, userId, contestId)
	if err != nil {
		return pkg.HandlePgErr(err, op)
	}
	return nil
}

const GetParticipantQuery = `
SELECT user_id
FROM contest_user 
WHERE user_id=$1 AND contest_id=$2
`

func (r *Repository) IsParticipant(ctx context.Context, contestId int32, userId int32) (bool, error) {
	const op = "Repository.IsParticipant"

	var id int32
	err := r.db.GetContext(ctx, &id, GetParticipantQuery, userId, contestId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, pkg.HandlePgErr(err, op)
	}

	return true, nil
}

const (
	ListParticipantsQuery = `
SELECT u.id, u.username, u.role, '' as hashed_pwd, u.created_at, u.updated_at
FROM contest_user cu
LEFT JOIN users u ON cu.user_id = u.id 
WHERE contest_id = $1 LIMIT $2 OFFSET $3
`
	CountParticipantsQuery = "SELECT COUNT(*) FROM contest_user WHERE contest_id = $1"
)

func (r *Repository) ListParticipants(ctx context.Context, filter models.ParticipantsFilter) (*models.UsersList, error) {
	const op = "Repository.ListParticipants"

	var participants []*models.User
	err := r.db.SelectContext(ctx, &participants,
		ListParticipantsQuery, filter.ContestId, filter.PageSize, filter.Offset())
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	var count int32
	err = r.db.GetContext(ctx, &count, CountParticipantsQuery, filter.ContestId)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	return &models.UsersList{
		Users: participants,
		Pagination: models.Pagination{
			Total: models.Total(count, filter.PageSize),
			Page:  filter.Page,
		},
	}, nil
}

const GetMonitorParticipantsQuery = `
SELECT cu.user_id, u.username, COUNT(DISTINCT s.problem_id) as solved_problems, 0 as penalty
FROM contest_user cu
         LEFT JOIN solutions s ON cu.user_id = s.user_id
    AND cu.contest_id = s.contest_id AND s.state = 200
         LEFT JOIN users u ON cu.user_id = u.id
WHERE cu.contest_id = $1
GROUP BY (cu.user_id, u.username)
`

const GetMonitorStatistics = `
SELECT cp.problem_id,
       COUNT(CASE WHEN s.state = 200 THEN 1 END)                   AS s_atts,
       COUNT(CASE WHEN s.state != 200 AND s.state != 1 THEN 1 END) AS uns_atts,
       COUNT(*)                                                    AS t_atts,
       cp.position
FROM contest_problem cp
         LEFT JOIN
     solutions s ON cp.problem_id = s.problem_id
         AND cp.contest_id = s.contest_id
WHERE cp.contest_id = $1
GROUP BY (cp.problem_id, cp.position)
ORDER BY cp.problem_id
`

const GetMonitorMainQuery = `
WITH UserSolutions AS (
    SELECT
        cu.user_id,
        cp.problem_id,
        cp.position,
        s.state,
        s.created_at,
        ROW_NUMBER() OVER (
            PARTITION BY cu.user_id,
                cp.problem_id
            ORDER BY
                s.created_at
            ) AS attempt_number,
        MIN(
        CASE WHEN s.state = 200 THEN s.created_at END
            ) OVER (
            PARTITION BY cu.user_id, cp.problem_id
            ) AS first_success_time
    FROM
        contest_user cu
            JOIN contest_problem cp ON cu.contest_id = cp.contest_id
            LEFT JOIN solutions s ON cu.user_id = s.user_id
            AND cp.problem_id = s.problem_id
            AND cu.contest_id = s.contest_id
    WHERE
            cu.contest_id = $1
),
FailedAttempts AS (
    SELECT
        user_id,
        problem_id,
        position,
        COUNT(
            CASE WHEN state != 200
                AND state != 1
                AND (
                first_success_time IS NULL
                OR created_at < first_success_time
            ) THEN 1 END
        ) AS failed_attempts,
        CASE WHEN BOOL_OR(state = 200) THEN 200 ELSE MAX(state) END AS final_state
    FROM
        UserSolutions
    GROUP BY
        user_id,
        problem_id,
		position
)
SELECT
    user_id,
    problem_id,
	position,
    COALESCE(failed_attempts, 0) AS f_atts,
    final_state as state
FROM
    FailedAttempts
WHERE
    user_id IS NOT NULL
  AND problem_id IS NOT NULL
ORDER BY
    user_id,
    problem_id
`

func (r *Repository) GetMonitor(ctx context.Context, contestId int32) (*models.Monitor, error) {
	const op = "Repository.GetMonitor"

	participants := make([]*models.ParticipantsStat, 0)
	err := r.db.SelectContext(ctx, &participants, GetMonitorParticipantsQuery, contestId)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	summary := make([]*models.ProblemStatSummary, 0)
	err = r.db.SelectContext(ctx, &summary, GetMonitorStatistics, contestId)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	m := make(map[int32][]*models.ProblemAttempts)

	rows, err := r.db.QueryxContext(ctx, GetMonitorMainQuery, contestId)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}
	defer rows.Close()

	for rows.Next() {
		var att models.ProblemAttempts
		err = rows.StructScan(&att)
		if err != nil {
			return nil, pkg.HandlePgErr(err, op)
		}

		if m[att.UserId] == nil {
			m[att.UserId] = make([]*models.ProblemAttempts, 0)
		}

		m[att.UserId] = append(m[att.UserId], &att)
	}

	for _, v := range participants {
		v.Attempts = m[v.UserId]
	}

	sort.Slice(participants, func(i, j int) bool {
		if participants[i].Solved != participants[j].Solved {
			return participants[i].Solved > participants[j].Solved
		}

		return participants[i].Penalty < participants[j].Penalty
	})

	monitor := &models.Monitor{
		Participants: participants,
		Summary:      summary,
	}

	return monitor, nil
}
