package budgeting

import (
	"context"
	"time"

	"encore.dev/beta/errs"

	"encore.app/config"
)

type Budget struct {
	ID              int64     `json:"id"`
	TeamID          int64     `json:"team_id"`
	LeverID         int64     `json:"lever_id"`
	PeriodStart     time.Time `json:"period_start"`
	PeriodEnd       time.Time `json:"period_end"`
	ConfiguredTotal float64   `json:"configured_total"`
}

type GetConfiguredBudgetsParams struct {
	TeamID      int64     `json:"team_id"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
}

type BudgetsResponse struct {
	Budgets []*Budget `json:"budgets"`
}

// GetConfiguredBudgets returns budgets for a team within a time period.
//
//encore:api public method=POST path=/budgeting/budgets/query
func GetConfiguredBudgets(ctx context.Context, p *GetConfiguredBudgetsParams) (*BudgetsResponse, error) {
	rows, err := db.Query(ctx, `
		SELECT id, team_id, lever_id, period_start, period_end, configured_total
		FROM budgets
		WHERE team_id = $1 AND period_start >= $2 AND period_end <= $3
		ORDER BY lever_id
	`, p.TeamID, p.PeriodStart, p.PeriodEnd)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to query budgets")
	}
	defer rows.Close()

	var budgets []*Budget
	for rows.Next() {
		var b Budget
		if err := rows.Scan(&b.ID, &b.TeamID, &b.LeverID, &b.PeriodStart, &b.PeriodEnd, &b.ConfiguredTotal); err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "failed to scan budget")
		}
		budgets = append(budgets, &b)
	}
	return &BudgetsResponse{Budgets: budgets}, nil
}

type RecordSpendParams struct {
	AssignmentID int64   `json:"assignment_id"`
	TeamID       int64   `json:"team_id"`
	LeverID      int64   `json:"lever_id"`
	Amount       float64 `json:"amount"`
}

// RecordSpend records an observed spend event against a budget.
//
//encore:api public method=POST path=/budgeting/spend
func RecordSpend(ctx context.Context, p *RecordSpendParams) error {
	_, err := db.Exec(ctx, `
		INSERT INTO spend_records (assignment_id, team_id, lever_id, amount)
		VALUES ($1, $2, $3, $4)
	`, p.AssignmentID, p.TeamID, p.LeverID, p.Amount)
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to record spend")
	}
	return nil
}

// SyncFromConfig pulls budget configs from the config service and upserts them locally.
//
//encore:api public method=POST path=/budgeting/sync
func SyncFromConfig(ctx context.Context) error {
	resp, err := config.ListBudgetConfigs(ctx, &config.ListBudgetConfigsParams{})
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to fetch budget configs")
	}

	for _, bc := range resp.BudgetConfigs {
		_, err := db.Exec(ctx, `
			INSERT INTO budgets (team_id, lever_id, period_start, period_end, configured_total)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (lever_id, period_start) DO UPDATE
			SET team_id = EXCLUDED.team_id, period_end = EXCLUDED.period_end, configured_total = EXCLUDED.configured_total
		`, bc.TeamID, bc.LeverID, bc.PeriodStart, bc.PeriodEnd, bc.ConfiguredTotal)
		if err != nil {
			return errs.WrapCode(err, errs.Internal, "failed to upsert budget")
		}
	}
	return nil
}
