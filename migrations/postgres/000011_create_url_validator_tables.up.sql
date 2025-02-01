CREATE TABLE url_validator_configs
(
    type       VARCHAR(32) PRIMARY KEY,
    protocols  TEXT[] NOT NULL,
    domains    TEXT[] NOT NULL,
    extensions TEXT[] NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
