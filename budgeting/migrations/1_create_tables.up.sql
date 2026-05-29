CREATE TABLE budgets (
    id BIGSERIAL PRIMARY KEY,
    team_id BIGINT NOT NULL,
    lever_id BIGINT NOT NULL,
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    configured_total NUMERIC(18,4) NOT NULL,
    UNIQUE (lever_id, period_start)
);

CREATE TABLE spend_records (
    id BIGSERIAL PRIMARY KEY,
    assignment_id BIGINT NOT NULL,
    team_id BIGINT NOT NULL,
    lever_id BIGINT NOT NULL,
    amount NUMERIC(18,4) NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX spend_team_period_idx ON spend_records (team_id, occurred_at);
