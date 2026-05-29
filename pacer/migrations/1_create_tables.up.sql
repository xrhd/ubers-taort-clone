CREATE TABLE pacer_state (
    team_id BIGINT NOT NULL,
    lever_id BIGINT NOT NULL,
    period_start TIMESTAMPTZ NOT NULL,
    observed_spend NUMERIC(18,4) NOT NULL DEFAULT 0,
    predicted_liability NUMERIC(18,4) NOT NULL DEFAULT 0,
    last_updated TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (team_id, lever_id, period_start)
);
