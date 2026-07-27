-- +goose Up

CREATE TYPE scan_status AS ENUM (
    'queued',
    'running',
    'completed',
    'failed',
    'cancelled'
);

CREATE TABLE scans (
    id UUID PRIMARY KEY,
    url TEXT NOT NULL CHECK (char_length(url) <= 2048),
    status scan_status NOT NULL DEFAULT 'queued',
    error_message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX scans_status_created_at_idx
    ON scans (status, created_at DESC);

-- +goose Down

DROP TABLE IF EXISTS scans;
DROP TYPE IF EXISTS scan_status;
