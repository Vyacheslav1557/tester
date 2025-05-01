package repository

import (
	"context"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/pkg"
)

const (
	// state=5 - AC
	ReadStatisticsQuery = `
SELECT t.id                                    as task_id,
       t.position,
       COUNT(*)                                as total,
       COUNT(CASE WHEN s.state = 5 THEN 1 END) as success
FROM tasks t LEFT JOIN solutions s ON t.id = s.task_id
WHERE t.contest_id = $1
GROUP BY t.id, t.position
ORDER BY t.position;
`

	SolutionsQuery = `
WITH RankedSolutions AS (
	SELECT
		s.id,

		s.participant_id,
		p2.name    as participant_name,

		s.state,
		s.score,
		s.penalty,
		s.time_stat,
		s.memory_stat,
		s.language,

		s.task_id,
		t.position as task_position,
		p.title    as task_title,

		t.contest_id,
		c.title as contest_title,

		s.updated_at,
		s.created_at,
		 ROW_NUMBER() OVER (
			PARTITION BY s.task_id, s.participant_id 
			ORDER BY s.score DESC, s.created_at
		) as rn
	FROM solutions s
			 LEFT JOIN tasks t ON s.task_id = t.id
			 LEFT JOIN problems p ON t.problem_id = p.id
			 LEFT JOIN contests c ON t.contest_id = c.id
			 LEFT JOIN participants p2 on s.participant_id = p2.id
	WHERE t.contest_id = $1
)
SELECT
	rs.id,

	rs.participant_id,
	rs.participant_name,

	rs.state,
	rs.score,
	rs.penalty,
	rs.time_stat,
	rs.memory_stat,
	rs.language,

	rs.task_id,
	rs.task_position,
	rs.task_title,

	rs.contest_id,
	rs.contest_title,

	rs.updated_at,
	rs.created_at
FROM RankedSolutions rs
WHERE rs.rn = 1`

	ParticipantsQuery = `
WITH Attempts AS (
    SELECT
        s.participant_id,
        s.task_id,
        COUNT(*) FILTER (WHERE s.state != 5 AND s.created_at < (
            SELECT MIN(s2.created_at)
            FROM solutions s2
            WHERE s2.participant_id = s.participant_id
              AND s2.task_id = s.task_id
              AND s2.state = 5
        )) as failed_attempts,
        MIN(CASE WHEN s.state = 5 THEN s.penalty END) as success_penalty
    FROM solutions s JOIN tasks t ON t.id = s.task_id
    WHERE t.contest_id = $1
    GROUP BY s.participant_id, s.task_id
)
SELECT
    p.id,
    p.name,
    COUNT(DISTINCT CASE WHEN a.success_penalty IS NOT NULL THEN a.task_id END) as solved_in_total,
    COALESCE(SUM(a.failed_attempts), 0) * $2 + COALESCE(SUM(a.success_penalty), 0) as penalty_in_total
FROM participants p LEFT JOIN Attempts a ON a.participant_id = p.id
WHERE p.contest_id = $1
GROUP BY p.id, p.name
`
)

func (r *Repository) GetMonitor(ctx context.Context, contestId int32, penalty int32) (*models.Monitor, error) {
	const op = "Repository.GetMonitor"

	rows, err := r.db.QueryxContext(ctx, ReadStatisticsQuery, contestId)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}
	defer rows.Close()

	var monitor models.Monitor
	for rows.Next() {
		var stat models.ProblemStatSummary
		err = rows.StructScan(&stat)
		if err != nil {
			return nil, pkg.HandlePgErr(err, op)
		}
		monitor.Summary = append(monitor.Summary, &stat)
	}

	var solutions []*models.SolutionsListItem
	err = r.db.SelectContext(ctx, &solutions, SolutionsQuery, contestId)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	rows3, err := r.db.QueryxContext(ctx, ParticipantsQuery, contestId, penalty)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}
	defer rows3.Close()

	solutionsMap := make(map[int32][]*models.SolutionsListItem)
	for _, solution := range solutions {
		solutionsMap[solution.ParticipantId] = append(solutionsMap[solution.ParticipantId], solution)
	}

	for rows3.Next() {
		var stat models.ParticipantsStat
		err = rows3.StructScan(&stat)
		if err != nil {
			return nil, pkg.HandlePgErr(err, op)
		}

		if sols, ok := solutionsMap[stat.Id]; ok {
			stat.Solutions = sols
		}

		monitor.Participants = append(monitor.Participants, &stat)
	}

	return &monitor, nil
}
