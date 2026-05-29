package pacer

import (
	"context"
	"time"

	"encore.dev/pubsub"

	"encore.app/events"
)

// Subscribe to spend events from domain service to update observed spend.
var _ = pubsub.NewSubscription(
	events.SpendEventsTopic, "consume-spend",
	pubsub.SubscriptionConfig[*events.SpendEvent]{
		Handler: handleSpendEvent,
	},
)

func handleSpendEvent(ctx context.Context, event *events.SpendEvent) error {
	return UpdateSpend(ctx, &UpdateSpendParams{
		TeamID:      event.TeamID,
		LeverID:     event.LeverID,
		PeriodStart: currentPeriodStart(),
		Amount:      event.Amount,
		Kind:        "observed",
	})
}

// currentPeriodStart returns the start of the current quarter.
func currentPeriodStart() time.Time {
	now := time.Now().UTC()
	quarter := (int(now.Month()) - 1) / 3
	return time.Date(now.Year(), time.Month(quarter*3+1), 1, 0, 0, 0, 0, time.UTC)
}
