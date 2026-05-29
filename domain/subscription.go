package domain

import (
	"context"

	"encore.dev/pubsub"

	"encore.app/events"
)

// Subscribe to assignment events from orchestrator.
var _ = pubsub.NewSubscription(
	events.AssignmentTopic, "apply-assignment",
	pubsub.SubscriptionConfig[*events.AssignmentMessage]{
		Handler: handleAssignment,
	},
)

func handleAssignment(ctx context.Context, msg *events.AssignmentMessage) error {
	// Insert assignment idempotently.
	var assignmentID int64
	err := db.QueryRow(ctx, `
		INSERT INTO assignments (run_id, user_id, lever_id, team_id, predicted_cost, predicted_roi, status)
		VALUES ($1, $2, $3, $4, $5, $6, 'applied')
		ON CONFLICT (run_id, user_id) DO UPDATE SET status = assignments.status
		RETURNING id
	`, msg.RunID, msg.UserID, msg.LeverID, msg.TeamID, msg.PredictedCost, msg.PredictedROI).Scan(&assignmentID)
	if err != nil {
		return err
	}

	// Record event idempotently.
	_, _ = db.Exec(ctx, `
		INSERT INTO assignment_events (assignment_id, kind)
		SELECT $1, 'applied'
		WHERE NOT EXISTS (SELECT 1 FROM assignment_events WHERE assignment_id = $1 AND kind = 'applied')
	`, assignmentID)

	// Publish spend event.
	_, err = events.SpendEventsTopic.Publish(ctx, &events.SpendEvent{
		AssignmentID: assignmentID,
		TeamID:       msg.TeamID,
		LeverID:      msg.LeverID,
		Amount:       msg.PredictedCost,
	})
	return err
}
