package optimizer

import "sort"

// Solver defines the interface for MKP solvers.
// The greedy implementation can be swapped for CP-SAT or branch-and-bound.
type Solver interface {
	Solve(candidates []Candidate, leverBudgets []LeverBudget) ([]Assignment, float64, float64)
}

// GreedySolver implements a greedy MKP solver that assigns users to levers
// in descending order of ROI/cost ratio, respecting per-lever budget caps
// and ensuring each user is assigned at most once.
type GreedySolver struct{}

func (s *GreedySolver) Solve(candidates []Candidate, leverBudgets []LeverBudget) ([]Assignment, float64, float64) {
	if len(candidates) == 0 || len(leverBudgets) == 0 {
		return nil, 0, 0
	}

	// Build budget map: leverID -> available.
	budgetMap := make(map[int64]float64, len(leverBudgets))
	for _, lb := range leverBudgets {
		budgetMap[lb.LeverID] = lb.Available
	}

	// Sort candidates by ROI/cost ratio descending.
	type rankedCandidate struct {
		Candidate
		ratio float64
	}
	ranked := make([]rankedCandidate, 0, len(candidates))
	for _, c := range candidates {
		if c.PredictedCost <= 0 {
			continue
		}
		if _, ok := budgetMap[c.LeverID]; !ok {
			continue
		}
		ranked = append(ranked, rankedCandidate{
			Candidate: c,
			ratio:     c.PredictedROI / c.PredictedCost,
		})
	}

	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].ratio > ranked[j].ratio
	})

	assigned := make(map[int64]bool)   // userID -> already assigned?
	remaining := make(map[int64]float64) // leverID -> remaining budget
	for k, v := range budgetMap {
		remaining[k] = v
	}

	var assignments []Assignment
	var totalCost, totalROI float64

	for _, rc := range ranked {
		if assigned[rc.UserID] {
			continue
		}
		rem := remaining[rc.LeverID]
		if rc.PredictedCost > rem {
			continue
		}
		assignments = append(assignments, Assignment{
			UserID:        rc.UserID,
			LeverID:       rc.LeverID,
			PredictedCost: rc.PredictedCost,
			PredictedROI:  rc.PredictedROI,
		})
		assigned[rc.UserID] = true
		remaining[rc.LeverID] = rem - rc.PredictedCost
		totalCost += rc.PredictedCost
		totalROI += rc.PredictedROI
	}

	return assignments, totalCost, totalROI
}
