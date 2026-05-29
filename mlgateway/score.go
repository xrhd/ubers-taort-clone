package mlgateway

import (
	"context"
	"encoding/json"

	"encore.dev/beta/errs"
)

type Opportunity struct {
	UserID  int64 `json:"user_id"`
	LeverID int64 `json:"lever_id"`
}

type Prediction struct {
	UserID        int64           `json:"user_id"`
	LeverID       int64           `json:"lever_id"`
	PredictedCost float64         `json:"predicted_cost"`
	PredictedROI  float64         `json:"predicted_roi"`
	UpliftVector  json.RawMessage `json:"uplift_vector"`
}

type ScoreRequest struct {
	RunID         int64          `json:"run_id"`
	Opportunities []*Opportunity `json:"opportunities"`
}

type ScoreResponse struct {
	Predictions []*Prediction `json:"predictions"`
}

// ScoreOpportunities generates predictions for a batch of (user, lever) pairs.
// Uses synthetic deterministic scoring and persists results for audit.
//
//encore:api public method=POST path=/mlgateway/score
func ScoreOpportunities(ctx context.Context, req *ScoreRequest) (*ScoreResponse, error) {
	// Load current model weights.
	var weightsJSON json.RawMessage
	var modelVersionID int64
	err := db.QueryRow(ctx, `
		SELECT id, weights FROM model_versions
		WHERE name = 'default'
		ORDER BY id DESC LIMIT 1
	`).Scan(&modelVersionID, &weightsJSON)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to load model version")
	}

	var weights map[string]float64
	if err := json.Unmarshal(weightsJSON, &weights); err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to parse model weights")
	}

	predictions := make([]*Prediction, 0, len(req.Opportunities))
	for _, opp := range req.Opportunities {
		cost, roi, uplift := syntheticPredict(opp.UserID, opp.LeverID, weights)

		// Persist prediction for audit trail.
		_, err := db.Exec(ctx, `
			INSERT INTO predictions (run_id, user_id, lever_id, predicted_cost, predicted_roi, uplift_vector, model_version_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, req.RunID, opp.UserID, opp.LeverID, cost, roi, uplift, modelVersionID)
		if err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "failed to persist prediction")
		}

		predictions = append(predictions, &Prediction{
			UserID:        opp.UserID,
			LeverID:       opp.LeverID,
			PredictedCost: cost,
			PredictedROI:  roi,
			UpliftVector:  uplift,
		})
	}

	return &ScoreResponse{Predictions: predictions}, nil
}
