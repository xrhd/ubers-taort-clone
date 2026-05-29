package segmentation

import (
	"context"
	"encoding/json"

	"encore.dev/beta/errs"

	"encore.app/config"
)

type GetEligibleUsersResponse struct {
	Users []*User `json:"users"`
}

// GetEligibleUsers resolves a cohort predicate into matching users using JSONB containment.
//
//encore:api public method=GET path=/segmentation/cohorts/:cohortID/users
func GetEligibleUsers(ctx context.Context, cohortID int64) (*GetEligibleUsersResponse, error) {
	cohort, err := config.GetCohort(ctx, cohortID)
	if err != nil {
		return nil, errs.WrapCode(err, errs.NotFound, "cohort not found")
	}

	// Validate that the predicate is a valid JSON object.
	var pred map[string]interface{}
	if err := json.Unmarshal(cohort.Predicate, &pred); err != nil {
		return nil, errs.WrapCode(err, errs.InvalidArgument, "invalid cohort predicate")
	}

	rows, err := db.Query(ctx, `
		SELECT id, name, segment_attrs, created_at
		FROM users
		WHERE segment_attrs @> $1::jsonb
		ORDER BY id
	`, cohort.Predicate)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to query eligible users")
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.SegmentAttrs, &u.CreatedAt); err != nil {
			return nil, errs.WrapCode(err, errs.Internal, "failed to scan user")
		}
		users = append(users, &u)
	}
	return &GetEligibleUsersResponse{Users: users}, nil
}
