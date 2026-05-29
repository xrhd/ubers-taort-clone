package orchestrator

import (
	"context"
	"time"

	"encore.dev/beta/errs"
	"encore.dev/cron"
)

type TriggerRunParams struct {
	TeamID   int64 `json:"team_id"`
	CohortID int64 `json:"cohort_id"`
}

type TriggerRunResponse struct {
	RunID int64 `json:"run_id"`
}

// TriggerRun initiates a new targeting run for a team and cohort.
//
//encore:api public method=POST path=/orchestrator/runs
func TriggerRun(ctx context.Context, p *TriggerRunParams) (*TriggerRunResponse, error) {
	var runID int64
	err := db.QueryRow(ctx, `
		INSERT INTO targeting_runs (team_id, cohort_id, status)
		VALUES ($1, $2, 'running')
		RETURNING id
	`, p.TeamID, p.CohortID).Scan(&runID)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to create targeting run")
	}

	// Execute the pipeline asynchronously-ish (in the same request for simplicity).
	err = executeRun(ctx, runID, p.TeamID, p.CohortID)
	if err != nil {
		// Mark run as failed.
		_, _ = db.Exec(ctx, `
			UPDATE targeting_runs SET status = 'failed', error = $2, finished_at = NOW()
			WHERE id = $1
		`, runID, err.Error())
		return nil, err
	}

	return &TriggerRunResponse{RunID: runID}, nil
}

type RunStatus struct {
	ID         int64      `json:"id"`
	TeamID     int64      `json:"team_id"`
	CohortID   int64      `json:"cohort_id"`
	Status     string     `json:"status"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	Error      *string    `json:"error,omitempty"`
	Steps      []*RunStep `json:"steps"`
}

type RunStep struct {
	Step       string     `json:"step"`
	Status     string     `json:"status"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

// GetRunStatus returns the status and steps of a targeting run.
//
//encore:api public method=GET path=/orchestrator/runs/:id
func GetRunStatus(ctx context.Context, id int64) (*RunStatus, error) {
	var rs RunStatus
	err := db.QueryRow(ctx, `
		SELECT id, team_id, cohort_id, status, started_at, finished_at, error
		FROM targeting_runs WHERE id = $1
	`, id).Scan(&rs.ID, &rs.TeamID, &rs.CohortID, &rs.Status, &rs.StartedAt, &rs.FinishedAt, &rs.Error)
	if err != nil {
		return nil, errs.WrapCode(err, errs.NotFound, "run not found")
	}

	rows, err := db.Query(ctx, `
		SELECT step, status, started_at, finished_at
		FROM run_steps WHERE run_id = $1 ORDER BY id
	`, id)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to query run steps")
	}
	defer rows.Close()

	for rows.Next() {
		var s RunStep
		if err := rows.Scan(&s.Step, &s.Status, &s.StartedAt, &s.FinishedAt); err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "failed to scan run step")
		}
		rs.Steps = append(rs.Steps, &s)
	}
	return &rs, nil
}

// TriggerCron is called by the cron job. Uses default team/cohort IDs (1, 1).
//
//encore:api private method=POST path=/orchestrator/runs/cron
func TriggerCron(ctx context.Context) error {
	_, err := TriggerRun(ctx, &TriggerRunParams{TeamID: 1, CohortID: 1})
	return err
}

var _ = cron.NewJob("trigger-targeting-run", cron.JobConfig{
	Title:    "Trigger TAROT targeting run",
	Endpoint: TriggerCron,
	Every:    1 * cron.Hour,
})
