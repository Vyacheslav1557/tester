package repository

import (
	"context"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/pkg"
)

const CreateTaskQuery = `INSERT INTO tasks (problem_id, contest_id, position)
VALUES ($1, $2, COALESCE((SELECT MAX(position) FROM tasks WHERE contest_id = $2), 0) + 1)
RETURNING id
`

func (r *Repository) CreateTask(ctx context.Context, contestId int32, problemId int32) (int32, error) {
	const op = "Repository.AddTask"

	rows, err := r.db.QueryxContext(ctx, CreateTaskQuery, problemId, contestId)
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

const DeleteTaskQuery = "DELETE FROM tasks WHERE id=$1"

func (r *Repository) DeleteTask(ctx context.Context, taskId int32) error {
	const op = "Repository.DeleteTask"

	_, err := r.db.ExecContext(ctx, DeleteTaskQuery, taskId)
	if err != nil {
		return pkg.HandlePgErr(err, op)
	}
	return nil
}

const GetTasksQuery = `SELECT tasks.id,
       problem_id,
       contest_id,
       position,
       title,
       memory_limit,
       time_limit,
       tasks.created_at,
       tasks.updated_at
FROM tasks
         INNER JOIN problems ON tasks.problem_id = problems.id
WHERE contest_id = $1 ORDER BY position`

func (r *Repository) GetTasks(ctx context.Context, contestId int32) ([]*models.TasksListItem, error) {
	const op = "Repository.ReadTasks"

	var tasks []*models.TasksListItem
	err := r.db.SelectContext(ctx, &tasks, GetTasksQuery, contestId)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}
	return tasks, nil
}

const (
	GetTaskQuery = `
        SELECT 
            t.id,
            t.position,
            p.title,
            p.time_limit,
            p.memory_limit,
            t.problem_id,
            t.contest_id,
            p.legend_html,
            p.input_format_html,
            p.output_format_html,
            p.notes_html,
            p.scoring_html,
            t.created_at,
            t.updated_at
        FROM tasks t
        LEFT JOIN problems p ON t.problem_id = p.id
        WHERE t.id = ?
    `
)

func (r *Repository) GetTask(ctx context.Context, id int32) (*models.Task, error) {
	const op = "Repository.ReadTask"

	query := r.db.Rebind(GetTaskQuery)
	var task models.Task
	err := r.db.GetContext(ctx, &task, query, id)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	return &task, nil
}
