// Package events defines shared Pub/Sub topics to avoid import cycles between services.
package events

import "encore.dev/pubsub"

// AssignmentMessage is published by orchestrator for each optimized assignment.
type AssignmentMessage struct {
	RunID         int64   `json:"run_id"`
	UserID        int64   `json:"user_id"`
	LeverID       int64   `json:"lever_id"`
	TeamID        int64   `json:"team_id"`
	PredictedCost float64 `json:"predicted_cost"`
	PredictedROI  float64 `json:"predicted_roi"`
}

// AssignmentTopic broadcasts optimized assignments to domain services.
var AssignmentTopic = pubsub.NewTopic[*AssignmentMessage]("targeting-assignments", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})

// SpendEvent is published by domain when an assignment's spend is observed.
type SpendEvent struct {
	AssignmentID int64   `json:"assignment_id"`
	TeamID       int64   `json:"team_id"`
	LeverID      int64   `json:"lever_id"`
	Amount       float64 `json:"amount"`
}

// SpendEventsTopic broadcasts observed spend to downstream consumers (pacer).
var SpendEventsTopic = pubsub.NewTopic[*SpendEvent]("spend-events", pubsub.TopicConfig{
	DeliveryGuarantee: pubsub.AtLeastOnce,
})
