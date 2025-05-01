package repository

import (
	"context"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/pkg"
)

const GetParticipantIdQuery = "SELECT id FROM participants WHERE user_id=$1 AND contest_id=$2 LIMIT 1"

func (r *Repository) GetParticipantId(ctx context.Context, contestId int32, userId int32) (int32, error) {
	const op = "Repository.GetParticipantId"

	var participantId int32
	err := r.db.GetContext(ctx, &participantId, GetParticipantIdQuery, userId, contestId)
	if err != nil {
		return 0, pkg.HandlePgErr(err, op)
	}

	return participantId, nil
}

const GetParticipantId2Query = "SELECT p.id FROM participants p JOIN tasks t ON p.contest_id=t.contest_id WHERE user_id=$1 AND t.id=$2 LIMIT 1"

func (r *Repository) GetParticipantId2(ctx context.Context, taskId int32, userId int32) (int32, error) {
	const op = "Repository.GetParticipantId2"

	var participantId int32
	err := r.db.GetContext(ctx, &participantId, GetParticipantId2Query, userId, taskId)
	if err != nil {
		return 0, pkg.HandlePgErr(err, op)
	}

	return participantId, nil
}

const GetParticipantId3Query = "SELECT participant_id FROM solutions WHERE id=$1 LIMIT 1"

func (r *Repository) GetParticipantId3(ctx context.Context, solutionId int32) (int32, error) {
	const op = "Repository.GetParticipantId3"

	var participantId int32
	err := r.db.GetContext(ctx, &participantId, GetParticipantId3Query, solutionId)
	if err != nil {
		return 0, pkg.HandlePgErr(err, op)
	}

	return participantId, nil
}

const CreateParticipantQuery = "INSERT INTO participants (user_id, contest_id, name) VALUES ($1, $2, $3) RETURNING id"

func (r *Repository) CreateParticipant(ctx context.Context, contestId int32, userId int32) (int32, error) {
	const op = "Repository.CreateParticipant"

	name := ""
	rows, err := r.db.QueryxContext(ctx, CreateParticipantQuery, userId, contestId, name)
	if err != nil {
		return 0, pkg.HandlePgErr(err, op)
	}
	defer rows.Close()
	var id int32
	rows.Next()
	err = rows.Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

const DeleteParticipantQuery = "DELETE FROM participants WHERE id=$1"

const (
	UpdateParticipantQuery = "UPDATE participants SET name = COALESCE($1, name) WHERE id = $2"
)

func (r *Repository) UpdateParticipant(ctx context.Context, id int32, participantUpdate models.ParticipantUpdate) error {
	const op = "Repository.UpdateParticipant"

	_, err := r.db.ExecContext(ctx, UpdateParticipantQuery, participantUpdate.Name, id)
	if err != nil {
		return pkg.HandlePgErr(err, op)
	}

	return nil
}

func (r *Repository) DeleteParticipant(ctx context.Context, participantId int32) error {
	const op = "Repository.DeleteParticipant"

	_, err := r.db.ExecContext(ctx, DeleteParticipantQuery, participantId)
	if err != nil {
		return pkg.HandlePgErr(err, op)
	}
	return nil
}

const (
	ReadParticipantsListQuery = `SELECT id, user_id, name, created_at, updated_at FROM participants WHERE contest_id = $1 LIMIT $2 OFFSET $3`
	CountParticipantsQuery    = "SELECT COUNT(*) FROM participants WHERE contest_id = $1"
)

func (r *Repository) ListParticipants(ctx context.Context, filter models.ParticipantsFilter) (*models.ParticipantsList, error) {
	const op = "Repository.ReadParticipants"

	var participants []*models.ParticipantsListItem
	err := r.db.SelectContext(ctx, &participants,
		ReadParticipantsListQuery, filter.ContestId, filter.PageSize, filter.Offset())
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	var count int32
	err = r.db.GetContext(ctx, &count, CountParticipantsQuery, filter.ContestId)
	if err != nil {
		return nil, pkg.HandlePgErr(err, op)
	}

	return &models.ParticipantsList{
		Participants: participants,
		Pagination: models.Pagination{
			Total: models.Total(count, filter.PageSize),
			Page:  filter.Page,
		},
	}, nil
}
