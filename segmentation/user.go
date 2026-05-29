package segmentation

import (
	"context"
	"encoding/json"
	"time"

	"encore.dev/beta/errs"
)

type User struct {
	ID           int64           `json:"id"`
	Name         string          `json:"name"`
	SegmentAttrs json.RawMessage `json:"segment_attrs"`
	CreatedAt    time.Time       `json:"created_at"`
}

type CreateUserParams struct {
	Name         string          `json:"name"`
	SegmentAttrs json.RawMessage `json:"segment_attrs"`
}

//encore:api public method=POST path=/segmentation/users
func CreateUser(ctx context.Context, p *CreateUserParams) (*User, error) {
	attrs := p.SegmentAttrs
	if len(attrs) == 0 {
		attrs = json.RawMessage(`{}`)
	}
	var u User
	err := db.QueryRow(ctx, `
		INSERT INTO users (name, segment_attrs)
		VALUES ($1, $2)
		RETURNING id, name, segment_attrs, created_at
	`, p.Name, attrs).Scan(&u.ID, &u.Name, &u.SegmentAttrs, &u.CreatedAt)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to create user")
	}
	return &u, nil
}

type ListUsersResponse struct {
	Users []*User `json:"users"`
}

//encore:api public method=GET path=/segmentation/users
func ListUsers(ctx context.Context) (*ListUsersResponse, error) {
	rows, err := db.Query(ctx, `
		SELECT id, name, segment_attrs, created_at
		FROM users ORDER BY id
	`)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "failed to list users")
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
	return &ListUsersResponse{Users: users}, nil
}
