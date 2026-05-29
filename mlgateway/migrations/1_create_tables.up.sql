CREATE TABLE model_versions (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    version TEXT NOT NULL,
    weights JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (name, version)
);

INSERT INTO model_versions (name, version, weights)
VALUES ('default', 'v1', '{"rides_trips":1.0,"eats_orders":0.8}');

CREATE TABLE predictions (
    id BIGSERIAL PRIMARY KEY,
    run_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    lever_id BIGINT NOT NULL,
    predicted_cost NUMERIC(18,4) NOT NULL,
    predicted_roi NUMERIC(18,6) NOT NULL,
    uplift_vector JSONB NOT NULL,
    model_version_id BIGINT NOT NULL REFERENCES model_versions(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX predictions_run_idx ON predictions (run_id);
