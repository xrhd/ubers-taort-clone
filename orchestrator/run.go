package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"encore.dev/beta/errs"

	"encore.app/config"
	"encore.app/domain"
	"encore.app/mlgateway"
	"encore.app/optimizer"
	"encore.app/pacer"
	"encore.app/segmentation"
)

// executeRun runs the full TAROT targeting pipeline.
func executeRun(ctx context.Context, runID, teamID, cohortID int64) error {
	// Step 1: Segmentation — get eligible users.
	if err := logStep(ctx, runID, "segmentation", "started"); err != nil {
		return err
	}
	usersResp, err := segmentation.GetEligibleUsers(ctx, cohortID)
	if err != nil {
		_ = failStep(ctx, runID, "segmentation")
		return errs.WrapCode(err, errs.Internal, "segmentation failed")
	}
	_ = completeStep(ctx, runID, "segmentation")

	if len(usersResp.Users) == 0 {
		return completeRun(ctx, runID, 0, 0, 0, 0)
	}

	// Step 2: Get levers for team.
	leversResp, err := config.ListLevers(ctx, &config.ListLeversParams{TeamID: teamID})
	if err != nil {
		return errs.WrapCode(err, errs.Internal, "failed to list levers")
	}
	if len(leversResp.Levers) == 0 {
		return completeRun(ctx, runID, len(usersResp.Users), 0, 0, 0)
	}

	// Step 3: ML scoring — build Cartesian product and score.
	if err := logStep(ctx, runID, "mlgateway", "started"); err != nil {
		return err
	}
	opportunities := make([]*mlgateway.Opportunity, 0, len(usersResp.Users)*len(leversResp.Levers))
	for _, u := range usersResp.Users {
		for _, l := range leversResp.Levers {
			opportunities = append(opportunities, &mlgateway.Opportunity{
				UserID:  u.ID,
				LeverID: l.ID,
			})
		}
	}
	scoreResp, err := mlgateway.ScoreOpportunities(ctx, &mlgateway.ScoreRequest{
		RunID:         runID,
		Opportunities: opportunities,
	})
	if err != nil {
		_ = failStep(ctx, runID, "mlgateway")
		return errs.WrapCode(err, errs.Internal, "ML scoring failed")
	}
	_ = completeStep(ctx, runID, "mlgateway")

	// Step 4: Budget pacing — get available budgets.
	if err := logStep(ctx, runID, "pacer", "started"); err != nil {
		return err
	}
	now := time.Now().UTC()
	periodStart := time.Date(now.Year(), quarterStart(now.Month()), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 3, 0)

	capsResp, err := pacer.GetAvailableBudget(ctx, &pacer.GetAvailableBudgetParams{
		TeamID:      teamID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	})
	if err != nil {
		_ = failStep(ctx, runID, "pacer")
		return errs.WrapCode(err, errs.Internal, "pacer failed")
	}
	_ = completeStep(ctx, runID, "pacer")

	// Step 5: Optimization — solve MKP.
	if err := logStep(ctx, runID, "optimizer", "started"); err != nil {
		return err
	}
	candidates := make([]*optimizer.Candidate, 0, len(scoreResp.Predictions))
	for _, p := range scoreResp.Predictions {
		candidates = append(candidates, &optimizer.Candidate{
			UserID:        p.UserID,
			LeverID:       p.LeverID,
			PredictedCost: p.PredictedCost,
			PredictedROI:  p.PredictedROI,
		})
	}
	leverBudgets := make([]*optimizer.LeverBudget, 0, len(capsResp.Caps))
	for _, cap := range capsResp.Caps {
		leverBudgets = append(leverBudgets, &optimizer.LeverBudget{
			LeverID:   cap.LeverID,
			Available: cap.Available,
		})
	}
	solveResp, err := optimizer.Solve(ctx, &optimizer.SolveRequest{
		RunID:        runID,
		Candidates:   candidates,
		LeverBudgets: leverBudgets,
	})
	if err != nil {
		_ = failStep(ctx, runID, "optimizer")
		return errs.WrapCode(err, errs.Internal, "optimizer failed")
	}
	_ = completeStep(ctx, runID, "optimizer")

	// Step 6: Assign incentives via domain service.
	// Calls domain.AssignIncentive synchronously for reliable delivery.
	// Domain publishes spend events to pacer via Pub/Sub.
	if err := logStep(ctx, runID, "assign", "started"); err != nil {
		return err
	}

	// Find team for each lever.
	leverTeam := make(map[int64]int64)
	for _, l := range leversResp.Levers {
		leverTeam[l.ID] = l.TeamID
	}

	for _, a := range solveResp.Assignments {
		tID := leverTeam[a.LeverID]
		if tID == 0 {
			tID = teamID
		}
		_, err := domain.AssignIncentive(ctx, &domain.AssignIncentiveParams{
			RunID:         runID,
			UserID:        a.UserID,
			LeverID:       a.LeverID,
			TeamID:        tID,
			PredictedCost: a.PredictedCost,
			PredictedROI:  a.PredictedROI,
		})
		if err != nil {
			_ = failStep(ctx, runID, "assign")
			return errs.WrapCode(err, errs.Internal, "failed to assign incentive")
		}

		// Update pacer liability for the predicted cost.
		_ = pacer.UpdateSpend(ctx, &pacer.UpdateSpendParams{
			TeamID:      tID,
			LeverID:     a.LeverID,
			PeriodStart: periodStart,
			Amount:      a.PredictedCost,
			Kind:        "liability",
		})
	}
	_ = completeStep(ctx, runID, "assign")

	return completeRun(ctx, runID, len(usersResp.Users), len(solveResp.Assignments), solveResp.TotalCost, solveResp.TotalROI)
}

func completeRun(ctx context.Context, runID int64, users, assignments int, totalCost, totalROI float64) error {
	summary := map[string]interface{}{
		"users":       users,
		"assignments": assignments,
		"total_cost":  totalCost,
		"total_roi":   totalROI,
	}
	summaryJSON, _ := json.Marshal(summary)
	_, err := db.Exec(ctx, `
		UPDATE targeting_runs SET status = 'completed', finished_at = NOW(), summary = $2
		WHERE id = $1
	`, runID, summaryJSON)
	return err
}

func logStep(ctx context.Context, runID int64, step, status string) error {
	_, err := db.Exec(ctx, `
		INSERT INTO run_steps (run_id, step, status) VALUES ($1, $2, $3)
	`, runID, step, status)
	return err
}

func completeStep(ctx context.Context, runID int64, step string) error {
	_, err := db.Exec(ctx, `
		UPDATE run_steps SET status = 'completed', finished_at = NOW()
		WHERE run_id = $1 AND step = $2 AND status = 'started'
	`, runID, step)
	return err
}

func failStep(ctx context.Context, runID int64, step string) error {
	_, err := db.Exec(ctx, fmt.Sprintf(`
		UPDATE run_steps SET status = 'failed', finished_at = NOW()
		WHERE run_id = $1 AND step = $2 AND status = 'started'
	`), runID, step)
	return err
}

func quarterStart(m time.Month) time.Month {
	return time.Month(((int(m) - 1) / 3) * 3 + 1)
}
