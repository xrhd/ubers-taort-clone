package optimizer

import (
	"context"
)

type Candidate struct {
	UserID        int64   `json:"user_id"`
	LeverID       int64   `json:"lever_id"`
	PredictedCost float64 `json:"predicted_cost"`
	PredictedROI  float64 `json:"predicted_roi"`
}

type LeverBudget struct {
	LeverID   int64   `json:"lever_id"`
	Available float64 `json:"available"`
}

type Assignment struct {
	UserID        int64   `json:"user_id"`
	LeverID       int64   `json:"lever_id"`
	PredictedCost float64 `json:"predicted_cost"`
	PredictedROI  float64 `json:"predicted_roi"`
}

type SolveRequest struct {
	RunID        int64          `json:"run_id"`
	Candidates   []*Candidate   `json:"candidates"`
	LeverBudgets []*LeverBudget `json:"lever_budgets"`
}

type SolveResponse struct {
	Assignments []*Assignment `json:"assignments"`
	TotalCost   float64       `json:"total_cost"`
	TotalROI    float64       `json:"total_roi"`
}

// Solve runs the MKP optimizer on the given candidates and budget constraints.
//
//encore:api public method=POST path=/optimizer/solve
func Solve(ctx context.Context, req *SolveRequest) (*SolveResponse, error) {
	// Convert pointer slices to value slices for the solver.
	candidates := make([]Candidate, len(req.Candidates))
	for i, c := range req.Candidates {
		candidates[i] = *c
	}
	budgets := make([]LeverBudget, len(req.LeverBudgets))
	for i, lb := range req.LeverBudgets {
		budgets[i] = *lb
	}

	solver := &GreedySolver{}
	assignments, totalCost, totalROI := solver.Solve(candidates, budgets)

	result := make([]*Assignment, len(assignments))
	for i := range assignments {
		result[i] = &assignments[i]
	}

	return &SolveResponse{
		Assignments: result,
		TotalCost:   totalCost,
		TotalROI:    totalROI,
	}, nil
}
