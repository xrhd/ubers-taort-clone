package pacer

import (
	"context"
	"math"
	"time"

	"encore.dev/beta/errs"

	"encore.app/budgeting"
)

type BudgetCap struct {
	TeamID      int64   `json:"team_id"`
	LeverID     int64   `json:"lever_id"`
	Available   float64 `json:"available"`
	Configured  float64 `json:"configured"`
	Observed    float64 `json:"observed"`
	Liability   float64 `json:"predicted_liability"`
}

type GetAvailableBudgetParams struct {
	TeamID      int64     `json:"team_id"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
}

type GetAvailableBudgetResponse struct {
	Caps []*BudgetCap `json:"caps"`
}

// GetAvailableBudget computes available budget per lever for the current run.
// Formula: available = max(0, configured - observed - liability)
//
//encore:api public method=POST path=/pacer/available
func GetAvailableBudget(ctx context.Context, p *GetAvailableBudgetParams) (*GetAvailableBudgetResponse, error) {
	// Get configured budgets from budgeting service.
	budgets, err := budgeting.GetConfiguredBudgets(ctx, &budgeting.GetConfiguredBudgetsParams{
		TeamID:      p.TeamID,
		PeriodStart: p.PeriodStart,
		PeriodEnd:   p.PeriodEnd,
	})
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to get configured budgets")
	}

	var caps []*BudgetCap
	for _, b := range budgets.Budgets {
		var observed, liability float64
		_ = db.QueryRow(ctx, `
			SELECT COALESCE(observed_spend, 0), COALESCE(predicted_liability, 0)
			FROM pacer_state
			WHERE team_id = $1 AND lever_id = $2 AND period_start = $3
		`, b.TeamID, b.LeverID, b.PeriodStart).Scan(&observed, &liability)

		available := math.Max(0, b.ConfiguredTotal-observed-liability)
		caps = append(caps, &BudgetCap{
			TeamID:     b.TeamID,
			LeverID:    b.LeverID,
			Available:  available,
			Configured: b.ConfiguredTotal,
			Observed:   observed,
			Liability:  liability,
		})
	}
	return &GetAvailableBudgetResponse{Caps: caps}, nil
}

type UpdateSpendParams struct {
	TeamID      int64     `json:"team_id"`
	LeverID     int64     `json:"lever_id"`
	PeriodStart time.Time `json:"period_start"`
	Amount      float64   `json:"amount"`
	Kind        string    `json:"kind"` // "observed" or "liability"
}

// UpdateSpend updates the pacer state for a given team/lever/period.
//
//encore:api private method=POST path=/pacer/update-spend
func UpdateSpend(ctx context.Context, p *UpdateSpendParams) error {
	var column string
	switch p.Kind {
	case "observed":
		column = "observed_spend"
	case "liability":
		column = "predicted_liability"
	default:
		return errs.B().Code(errs.InvalidArgument).Msg("kind must be 'observed' or 'liability'").Err()
	}

	_, err := db.Exec(ctx, `
		INSERT INTO pacer_state (team_id, lever_id, period_start, `+column+`, last_updated)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (team_id, lever_id, period_start) DO UPDATE
		SET `+column+` = pacer_state.`+column+` + EXCLUDED.`+column+`,
		    last_updated = NOW()
	`, p.TeamID, p.LeverID, p.PeriodStart, p.Amount)
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to update pacer state")
	}
	return nil
}
