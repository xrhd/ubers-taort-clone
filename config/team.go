package config

import (
	"context"
	"time"

	"encore.dev/beta/errs"
)

type Team struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	LineOfBusiness string    `json:"line_of_business"`
	CreatedAt      time.Time `json:"created_at"`
}

type CreateTeamParams struct {
	Name           string `json:"name"`
	LineOfBusiness string `json:"line_of_business"`
}

//encore:api public method=POST path=/config/teams
func CreateTeam(ctx context.Context, p *CreateTeamParams) (*Team, error) {
	var t Team
	err := db.QueryRow(ctx, `
		INSERT INTO teams (name, line_of_business)
		VALUES ($1, $2)
		RETURNING id, name, line_of_business, created_at
	`, p.Name, p.LineOfBusiness).Scan(&t.ID, &t.Name, &t.LineOfBusiness, &t.CreatedAt)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to create team")
	}
	return &t, nil
}

//encore:api public method=GET path=/config/teams/:id
func GetTeam(ctx context.Context, id int64) (*Team, error) {
	var t Team
	err := db.QueryRow(ctx, `
		SELECT id, name, line_of_business, created_at
		FROM teams WHERE id = $1
	`, id).Scan(&t.ID, &t.Name, &t.LineOfBusiness, &t.CreatedAt)
	if err != nil {
		return nil, errs.WrapCode(err, errs.NotFound, "team not found")
	}
	return &t, nil
}

type ListTeamsResponse struct {
	Teams []*Team `json:"teams"`
}

//encore:api public method=GET path=/config/teams
func ListTeams(ctx context.Context) (*ListTeamsResponse, error) {
	rows, err := db.Query(ctx, `
		SELECT id, name, line_of_business, created_at
		FROM teams ORDER BY id
	`)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to list teams")
	}
	defer rows.Close()

	var teams []*Team
	for rows.Next() {
		var t Team
		if err := rows.Scan(&t.ID, &t.Name, &t.LineOfBusiness, &t.CreatedAt); err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "failed to scan team")
		}
		teams = append(teams, &t)
	}
	return &ListTeamsResponse{Teams: teams}, nil
}
