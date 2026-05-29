package domain

import (
	"context"
	"time"

	"encore.dev/beta/errs"

	"encore.app/events"
)

type AssignmentRecord struct {
	ID            int64     `json:"id"`
	RunID         int64     `json:"run_id"`
	UserID        int64     `json:"user_id"`
	LeverID       int64     `json:"lever_id"`
	TeamID        int64     `json:"team_id"`
	PredictedCost float64   `json:"predicted_cost"`
	PredictedROI  float64   `json:"predicted_roi"`
	Status        string    `json:"status"`
	AssignedAt    time.Time `json:"assigned_at"`
}

type AssignIncentiveParams struct {
	RunID         int64   `json:"run_id"`
	UserID        int64   `json:"user_id"`
	LeverID       int64   `json:"lever_id"`
	TeamID        int64   `json:"team_id"`
	PredictedCost float64 `json:"predicted_cost"`
	PredictedROI  float64 `json:"predicted_roi"`
}

// AssignIncentive records an assignment and publishes spend.
//
//encore:api public method=POST path=/domain/assign
func AssignIncentive(ctx context.Context, p *AssignIncentiveParams) (*AssignmentRecord, error) {
	// Use INSERT ... ON CONFLICT to handle idempotent retries from Pub/Sub.
	var a AssignmentRecord
	err := db.QueryRow(ctx, `
		INSERT INTO assignments (run_id, user_id, lever_id, team_id, predicted_cost, predicted_roi, status)
		VALUES ($1, $2, $3, $4, $5, $6, 'applied')
		ON CONFLICT (run_id, user_id) DO UPDATE SET status = assignments.status
		RETURNING id, run_id, user_id, lever_id, team_id, predicted_cost, predicted_roi, status, assigned_at
	`, p.RunID, p.UserID, p.LeverID, p.TeamID, p.PredictedCost, p.PredictedROI).Scan(
		&a.ID, &a.RunID, &a.UserID, &a.LeverID, &a.TeamID,
		&a.PredictedCost, &a.PredictedROI, &a.Status, &a.AssignedAt,
	)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to record assignment")
	}

	// Record assignment event (idempotent: check if already exists).
	_, _ = db.Exec(ctx, `
		INSERT INTO assignment_events (assignment_id, kind)
		SELECT $1, 'applied'
		WHERE NOT EXISTS (SELECT 1 FROM assignment_events WHERE assignment_id = $1 AND kind = 'applied')
	`, a.ID)

	// Publish spend event (simulate observed spend = predicted cost).
	_, err = events.SpendEventsTopic.Publish(ctx, &events.SpendEvent{
		AssignmentID: a.ID,
		TeamID:       p.TeamID,
		LeverID:      p.LeverID,
		Amount:       p.PredictedCost,
	})
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to publish spend event")
	}

	return &a, nil
}

type ListAssignmentsParams struct {
	RunID int64 `query:"run_id"`
}

type ListAssignmentsResponse struct {
	Assignments []*AssignmentRecord `json:"assignments"`
}

//encore:api public method=GET path=/domain/assignments
func ListAssignments(ctx context.Context, p *ListAssignmentsParams) (*ListAssignmentsResponse, error) {
	rows, err := db.Query(ctx, `
		SELECT id, run_id, user_id, lever_id, team_id, predicted_cost, predicted_roi, status, assigned_at
		FROM assignments
		WHERE run_id = $1
		ORDER BY id
	`, p.RunID)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to list assignments")
	}
	defer rows.Close()

	var assignments []*AssignmentRecord
	for rows.Next() {
		var a AssignmentRecord
		if err := rows.Scan(&a.ID, &a.RunID, &a.UserID, &a.LeverID, &a.TeamID,
			&a.PredictedCost, &a.PredictedROI, &a.Status, &a.AssignedAt); err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "failed to scan assignment")
		}
		assignments = append(assignments, &a)
	}
	return &ListAssignmentsResponse{Assignments: assignments}, nil
}
