package config

import (
	"context"
	"encoding/json"
	"time"

	"encore.dev/beta/errs"
)

type CohortConfig struct {
	ID        int64           `json:"id"`
	Name      string          `json:"name"`
	Predicate json.RawMessage `json:"predicate"`
	CreatedAt time.Time       `json:"created_at"`
}

type CreateCohortParams struct {
	Name      string          `json:"name"`
	Predicate json.RawMessage `json:"predicate"`
}

//encore:api public method=POST path=/config/cohorts
func CreateCohort(ctx context.Context, p *CreateCohortParams) (*CohortConfig, error) {
	var c CohortConfig
	err := db.QueryRow(ctx, `
		INSERT INTO cohort_configs (name, predicate)
		VALUES ($1, $2)
		RETURNING id, name, predicate, created_at
	`, p.Name, p.Predicate).Scan(&c.ID, &c.Name, &c.Predicate, &c.CreatedAt)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to create cohort")
	}
	return &c, nil
}

//encore:api public method=GET path=/config/cohorts/:id
func GetCohort(ctx context.Context, id int64) (*CohortConfig, error) {
	var c CohortConfig
	err := db.QueryRow(ctx, `
		SELECT id, name, predicate, created_at
		FROM cohort_configs WHERE id = $1
	`, id).Scan(&c.ID, &c.Name, &c.Predicate, &c.CreatedAt)
	if err != nil {
		return nil, errs.WrapCode(err, errs.NotFound, "cohort not found")
	}
	return &c, nil
}

type ListCohortsResponse struct {
	Cohorts []*CohortConfig `json:"cohorts"`
}

//encore:api public method=GET path=/config/cohorts
func ListCohorts(ctx context.Context) (*ListCohortsResponse, error) {
	rows, err := db.Query(ctx, `
		SELECT id, name, predicate, created_at
		FROM cohort_configs ORDER BY id
	`)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to list cohorts")
	}
	defer rows.Close()

	var cohorts []*CohortConfig
	for rows.Next() {
		var c CohortConfig
		if err := rows.Scan(&c.ID, &c.Name, &c.Predicate, &c.CreatedAt); err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "failed to scan cohort")
		}
		cohorts = append(cohorts, &c)
	}
	return &ListCohortsResponse{Cohorts: cohorts}, nil
}
