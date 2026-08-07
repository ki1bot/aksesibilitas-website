-- +goose Up

CREATE TABLE password_reset_tokens (
    token_hash CHAR(64) PRIMARY KEY,
    user_id UUID NOT NULL UNIQUE
        REFERENCES users(id)
        ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX password_reset_tokens_expires_at_idx
    ON password_reset_tokens (expires_at);

CREATE TABLE rate_limits (
    key_hash CHAR(64) PRIMARY KEY,
    request_count BIGINT NOT NULL DEFAULT 1,
    window_started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX rate_limits_expires_at_idx
    ON rate_limits (expires_at);

-- +goose Down

DROP TABLE IF EXISTS rate_limits;
DROP TABLE IF EXISTS password_reset_tokens;