package config

import (
	"context"
	"encoding/json"
	"time"

	"encore.dev/beta/errs"
)

type Lever struct {
	ID                   int64           `json:"id"`
	TeamID               int64           `json:"team_id"`
	Name                 string          `json:"name"`
	Description          string          `json:"description"`
	EligibilityPredicate json.RawMessage `json:"eligibility_predicate"`
	CreatedAt            time.Time       `json:"created_at"`
}

type CreateLeverParams struct {
	TeamID               int64           `json:"team_id"`
	Name                 string          `json:"name"`
	Description          string          `json:"description"`
	EligibilityPredicate json.RawMessage `json:"eligibility_predicate"`
}

//encore:api public method=POST path=/config/levers
func CreateLever(ctx context.Context, p *CreateLeverParams) (*Lever, error) {
	pred := p.EligibilityPredicate
	if len(pred) == 0 {
		pred = json.RawMessage(`{}`)
	}
	var l Lever
	err := db.QueryRow(ctx, `
		INSERT INTO levers (team_id, name, description, eligibility_predicate)
		VALUES ($1, $2, $3, $4)
		RETURNING id, team_id, name, description, eligibility_predicate, created_at
	`, p.TeamID, p.Name, p.Description, pred).Scan(
		&l.ID, &l.TeamID, &l.Name, &l.Description, &l.EligibilityPredicate, &l.CreatedAt,
	)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to create lever")
	}
	return &l, nil
}

type GetLeverResponse = Lever

//encore:api public method=GET path=/config/levers/:id
func GetLever(ctx context.Context, id int64) (*Lever, error) {
	var l Lever
	err := db.QueryRow(ctx, `
		SELECT id, team_id, name, description, eligibility_predicate, created_at
		FROM levers WHERE id = $1
	`, id).Scan(&l.ID, &l.TeamID, &l.Name, &l.Description, &l.EligibilityPredicate, &l.CreatedAt)
	if err != nil {
		return nil, errs.WrapCode(err, errs.NotFound, "lever not found")
	}
	return &l, nil
}

type ListLeversParams struct {
	TeamID int64 `query:"team_id"`
}

type ListLeversResponse struct {
	Levers []*Lever `json:"levers"`
}

//encore:api public method=GET path=/config/levers
func ListLevers(ctx context.Context, p *ListLeversParams) (*ListLeversResponse, error) {
	query := `SELECT id, team_id, name, description, eligibility_predicate, created_at FROM levers`
	var args []interface{}
	if p.TeamID > 0 {
		query += ` WHERE team_id = $1`
		args = append(args, p.TeamID)
	}
	query += ` ORDER BY id`

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to list levers")
	}
	defer rows.Close()

	var levers []*Lever
	for rows.Next() {
		var l Lever
		if err := rows.Scan(&l.ID, &l.TeamID, &l.Name, &l.Description, &l.EligibilityPredicate, &l.CreatedAt); err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "failed to scan lever")
		}
		levers = append(levers, &l)
	}
	return &ListLeversResponse{Levers: levers}, nil
}
