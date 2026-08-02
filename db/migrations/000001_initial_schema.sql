-- +goose Up

CREATE TYPE scan_status AS ENUM (
    'queued',
    'running',
    'completed',
    'failed',
    'cancelled'
);

CREATE TYPE violation_impact AS ENUM (
    'critical',
    'serious',
    'moderate',
    'minor'
);

CREATE TYPE review_status AS ENUM (
    'pending',
    'passed',
    'failed',
    'not_applicable'
);

CREATE TYPE report_format AS ENUM (
    'json',
    'pdf'
);

CREATE TABLE users (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT users_name_length CHECK (
        char_length(name) BETWEEN 2 AND 100
    ),
    CONSTRAINT users_email_length CHECK (
        char_length(email) BETWEEN 3 AND 255
    )
);

CREATE UNIQUE INDEX users_email_unique_idx
    ON users (LOWER(email));

CREATE TABLE sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash CHAR(64) NOT NULL UNIQUE,
    csrf_hash CHAR(64) NOT NULL,
    user_agent TEXT NOT NULL DEFAULT '',
    ip_address VARCHAR(64) NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ NOT NULL,
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX sessions_user_id_idx
    ON sessions (user_id);

CREATE INDEX sessions_expires_at_idx
    ON sessions (expires_at);

CREATE TABLE projects (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(120) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT projects_name_length CHECK (
        char_length(name) BETWEEN 2 AND 120
    ),
    CONSTRAINT projects_description_length CHECK (
        char_length(description) <= 1000
    )
);

CREATE INDEX projects_owner_id_idx
    ON projects (owner_id);

CREATE TABLE project_members (
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL DEFAULT 'member',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id, user_id),
    CONSTRAINT project_members_role_check CHECK (
        role IN ('owner', 'member')
    )
);

CREATE INDEX project_members_user_id_idx
    ON project_members (user_id);

CREATE TABLE scans (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    status scan_status NOT NULL DEFAULT 'queued',
    page_title TEXT NOT NULL DEFAULT '',
    automated_score SMALLINT NOT NULL DEFAULT 0,
    critical_count INTEGER NOT NULL DEFAULT 0,
    serious_count INTEGER NOT NULL DEFAULT 0,
    moderate_count INTEGER NOT NULL DEFAULT 0,
    minor_count INTEGER NOT NULL DEFAULT 0,
    error_message TEXT NOT NULL DEFAULT '',
    duration_ms BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    CONSTRAINT scans_url_length CHECK (
        char_length(url) BETWEEN 1 AND 2048
    ),
    CONSTRAINT scans_score_range CHECK (
        automated_score BETWEEN 0 AND 100
    ),
    CONSTRAINT scans_count_non_negative CHECK (
        critical_count >= 0
        AND serious_count >= 0
        AND moderate_count >= 0
        AND minor_count >= 0
    )
);

CREATE INDEX scans_project_created_at_idx
    ON scans (project_id, created_at DESC);

CREATE INDEX scans_created_by_created_at_idx
    ON scans (created_by, created_at DESC);

CREATE INDEX scans_status_created_at_idx
    ON scans (status, created_at DESC);

CREATE TABLE scanned_pages (
    id UUID PRIMARY KEY,
    scan_id UUID NOT NULL UNIQUE REFERENCES scans(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    language VARCHAR(32) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE violations (
    id UUID PRIMARY KEY,
    scanned_page_id UUID NOT NULL REFERENCES scanned_pages(id) ON DELETE CASCADE,
    rule_id VARCHAR(150) NOT NULL,
    impact violation_impact NOT NULL,
    description TEXT NOT NULL,
    help TEXT NOT NULL,
    help_url TEXT NOT NULL DEFAULT '',
    tags TEXT[] NOT NULL DEFAULT '{}',
    review_status review_status NOT NULL DEFAULT 'pending',
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT violations_rule_id_length CHECK (
        char_length(rule_id) BETWEEN 1 AND 150
    ),
    CONSTRAINT violations_notes_length CHECK (
        char_length(notes) <= 5000
    ),
    UNIQUE (scanned_page_id, rule_id)
);

CREATE INDEX violations_page_impact_idx
    ON violations (scanned_page_id, impact);

CREATE TABLE violation_nodes (
    id UUID PRIMARY KEY,
    violation_id UUID NOT NULL REFERENCES violations(id) ON DELETE CASCADE,
    html TEXT NOT NULL,
    target TEXT[] NOT NULL DEFAULT '{}',
    failure_summary TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX violation_nodes_violation_id_idx
    ON violation_nodes (violation_id);

CREATE TABLE manual_reviews (
    id UUID PRIMARY KEY,
    scan_id UUID NOT NULL UNIQUE REFERENCES scans(id) ON DELETE CASCADE,
    status review_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE manual_review_items (
    id UUID PRIMARY KEY,
    manual_review_id UUID NOT NULL REFERENCES manual_reviews(id) ON DELETE CASCADE,
    criterion VARCHAR(200) NOT NULL,
    instruction TEXT NOT NULL,
    status review_status NOT NULL DEFAULT 'pending',
    notes TEXT NOT NULL DEFAULT '',
    position INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT manual_review_items_position_positive CHECK (
        position > 0
    ),
    CONSTRAINT manual_review_items_notes_length CHECK (
        char_length(notes) <= 5000
    ),
    UNIQUE (manual_review_id, position)
);

CREATE INDEX manual_review_items_review_id_idx
    ON manual_review_items (manual_review_id, position);

CREATE TABLE reports (
    id UUID PRIMARY KEY,
    scan_id UUID NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    format report_format NOT NULL,
    filename VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    content BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX reports_scan_id_created_at_idx
    ON reports (scan_id, created_at DESC);

CREATE TABLE api_keys (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    key_prefix VARCHAR(32) NOT NULL,
    key_hash CHAR(64) NOT NULL UNIQUE,
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX api_keys_user_id_idx
    ON api_keys (user_id);

CREATE TABLE activity_logs (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    entity_id UUID,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX activity_logs_user_created_at_idx
    ON activity_logs (user_id, created_at DESC);

CREATE INDEX activity_logs_project_created_at_idx
    ON activity_logs (project_id, created_at DESC);

-- +goose Down

DROP TABLE IF EXISTS activity_logs;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS reports;
DROP TABLE IF EXISTS manual_review_items;
DROP TABLE IF EXISTS manual_reviews;
DROP TABLE IF EXISTS violation_nodes;
DROP TABLE IF EXISTS violations;
DROP TABLE IF EXISTS scanned_pages;
DROP TABLE IF EXISTS scans;
DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS report_format;
DROP TYPE IF EXISTS review_status;
DROP TYPE IF EXISTS violation_impact;
DROP TYPE IF EXISTS scan_status;