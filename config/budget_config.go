package config

import (
	"context"
	"time"

	"encore.dev/beta/errs"
)

type BudgetConfig struct {
	ID              int64     `json:"id"`
	TeamID          int64     `json:"team_id"`
	LeverID         int64     `json:"lever_id"`
	PeriodStart     time.Time `json:"period_start"`
	PeriodEnd       time.Time `json:"period_end"`
	ConfiguredTotal float64   `json:"configured_total"`
}

type CreateBudgetConfigParams struct {
	TeamID          int64     `json:"team_id"`
	LeverID         int64     `json:"lever_id"`
	PeriodStart     time.Time `json:"period_start"`
	PeriodEnd       time.Time `json:"period_end"`
	ConfiguredTotal float64   `json:"configured_total"`
}

//encore:api public method=POST path=/config/budget-configs
func CreateBudgetConfig(ctx context.Context, p *CreateBudgetConfigParams) (*BudgetConfig, error) {
	var bc BudgetConfig
	err := db.QueryRow(ctx, `
		INSERT INTO budget_configs (team_id, lever_id, period_start, period_end, configured_total)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (lever_id, period_start) DO UPDATE
		SET team_id = EXCLUDED.team_id, period_end = EXCLUDED.period_end, configured_total = EXCLUDED.configured_total
		RETURNING id, team_id, lever_id, period_start, period_end, configured_total
	`, p.TeamID, p.LeverID, p.PeriodStart, p.PeriodEnd, p.ConfiguredTotal).Scan(
		&bc.ID, &bc.TeamID, &bc.LeverID, &bc.PeriodStart, &bc.PeriodEnd, &bc.ConfiguredTotal,
	)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to upsert budget config")
	}
	return &bc, nil
}

type ListBudgetConfigsParams struct {
	TeamID int64 `query:"team_id"`
}

type ListBudgetConfigsResponse struct {
	BudgetConfigs []*BudgetConfig `json:"budget_configs"`
}

//encore:api public method=GET path=/config/budget-configs
func ListBudgetConfigs(ctx context.Context, p *ListBudgetConfigsParams) (*ListBudgetConfigsResponse, error) {
	query := `SELECT id, team_id, lever_id, period_start, period_end, configured_total FROM budget_configs`
	var args []interface{}
	if p.TeamID > 0 {
		query += ` WHERE team_id = $1`
		args = append(args, p.TeamID)
	}
	query += ` ORDER BY id`

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to list budget configs")
	}
	defer rows.Close()

	var configs []*BudgetConfig
	for rows.Next() {
		var bc BudgetConfig
		if err := rows.Scan(&bc.ID, &bc.TeamID, &bc.LeverID, &bc.PeriodStart, &bc.PeriodEnd, &bc.ConfiguredTotal); err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "failed to scan budget config")
		}
		configs = append(configs, &bc)
	}
	return &ListBudgetConfigsResponse{BudgetConfigs: configs}, nil
}
