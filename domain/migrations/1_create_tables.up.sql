CREATE TABLE assignments (
    id BIGSERIAL PRIMARY KEY,
    run_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    lever_id BIGINT NOT NULL,
    team_id BIGINT NOT NULL,
    predicted_cost NUMERIC(18,4) NOT NULL,
    predicted_roi NUMERIC(18,6) NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (run_id, user_id)
);
CREATE INDEX assignments_run_idx ON assignments (run_id);
CREATE INDEX assignments_user_idx ON assignments (user_id);

CREATE TABLE assignment_events (
    id BIGSERIAL PRIMARY KEY,
    assignment_id BIGINT NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
    kind TEXT NOT NULL,
    detail JSONB NOT NULL DEFAULT '{}'::JSONB,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
