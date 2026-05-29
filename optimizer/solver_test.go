package optimizer

import (
	"testing"
)

func TestGreedySolver_Empty(t *testing.T) {
	s := &GreedySolver{}
	assignments, cost, roi := s.Solve(nil, nil)
	if len(assignments) != 0 || cost != 0 || roi != 0 {
		t.Fatalf("expected empty result, got %d assignments", len(assignments))
	}
}

func TestGreedySolver_SingleLever(t *testing.T) {
	s := &GreedySolver{}
	candidates := []Candidate{
		{UserID: 1, LeverID: 1, PredictedCost: 5, PredictedROI: 10},
		{UserID: 2, LeverID: 1, PredictedCost: 3, PredictedROI: 9},
		{UserID: 3, LeverID: 1, PredictedCost: 4, PredictedROI: 4},
	}
	budgets := []LeverBudget{{LeverID: 1, Available: 8}}

	assignments, totalCost, totalROI := s.Solve(candidates, budgets)

	// User 2 has ratio 3.0, User 1 has ratio 2.0, User 3 has ratio 1.0
	// Should pick User 2 (cost 3) then User 1 (cost 5) = total cost 8
	if len(assignments) != 2 {
		t.Fatalf("expected 2 assignments, got %d", len(assignments))
	}
	if totalCost != 8 {
		t.Errorf("expected total cost 8, got %f", totalCost)
	}
	if totalROI != 19 {
		t.Errorf("expected total ROI 19, got %f", totalROI)
	}
}

func TestGreedySolver_MultiLever(t *testing.T) {
	s := &GreedySolver{}
	candidates := []Candidate{
		{UserID: 1, LeverID: 1, PredictedCost: 5, PredictedROI: 15},
		{UserID: 1, LeverID: 2, PredictedCost: 3, PredictedROI: 6},
		{UserID: 2, LeverID: 1, PredictedCost: 4, PredictedROI: 8},
		{UserID: 2, LeverID: 2, PredictedCost: 2, PredictedROI: 5},
	}
	budgets := []LeverBudget{
		{LeverID: 1, Available: 10},
		{LeverID: 2, Available: 5},
	}

	assignments, _, _ := s.Solve(candidates, budgets)

	// User 1 lever 1: ratio 3.0 (best), User 2 lever 2: ratio 2.5, User 2 lever 1: ratio 2.0, User 1 lever 2: ratio 2.0
	// Pick User 1 -> lever 1. Then User 2 -> lever 2. Each user assigned once.
	if len(assignments) != 2 {
		t.Fatalf("expected 2 assignments, got %d", len(assignments))
	}

	userLevers := make(map[int64]int64)
	for _, a := range assignments {
		userLevers[a.UserID] = a.LeverID
	}
	if userLevers[1] != 1 {
		t.Errorf("expected user 1 assigned to lever 1, got lever %d", userLevers[1])
	}
	if userLevers[2] != 2 {
		t.Errorf("expected user 2 assigned to lever 2, got lever %d", userLevers[2])
	}
}

func TestGreedySolver_BudgetExhausted(t *testing.T) {
	s := &GreedySolver{}
	candidates := []Candidate{
		{UserID: 1, LeverID: 1, PredictedCost: 10, PredictedROI: 20},
		{UserID: 2, LeverID: 1, PredictedCost: 10, PredictedROI: 15},
	}
	budgets := []LeverBudget{{LeverID: 1, Available: 10}}

	assignments, totalCost, _ := s.Solve(candidates, budgets)

	if len(assignments) != 1 {
		t.Fatalf("expected 1 assignment, got %d", len(assignments))
	}
	if totalCost != 10 {
		t.Errorf("expected total cost 10, got %f", totalCost)
	}
	if assignments[0].UserID != 1 {
		t.Errorf("expected user 1 (higher ratio), got user %d", assignments[0].UserID)
	}
}

func TestGreedySolver_ZeroCostSkipped(t *testing.T) {
	s := &GreedySolver{}
	candidates := []Candidate{
		{UserID: 1, LeverID: 1, PredictedCost: 0, PredictedROI: 10},
		{UserID: 2, LeverID: 1, PredictedCost: 5, PredictedROI: 10},
	}
	budgets := []LeverBudget{{LeverID: 1, Available: 10}}

	assignments, _, _ := s.Solve(candidates, budgets)
	if len(assignments) != 1 {
		t.Fatalf("expected 1 assignment (zero cost skipped), got %d", len(assignments))
	}
	if assignments[0].UserID != 2 {
		t.Errorf("expected user 2, got user %d", assignments[0].UserID)
	}
}
