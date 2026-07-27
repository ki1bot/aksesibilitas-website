-- name: CreateScan :one
INSERT INTO scans (
    id,
    url
) VALUES (
    sqlc.arg(id),
    sqlc.arg(url)
)
RETURNING
    id,
    url,
    status,
    error_message,
    created_at,
    updated_at;

-- name: GetScan :one
SELECT
    id,
    url,
    status,
    error_message,
    created_at,
    updated_at
FROM scans
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: UpdateScanStatus :one
UPDATE scans
SET
    status = sqlc.arg(status),
    error_message = sqlc.arg(error_message),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING
    id,
    url,
    status,
    error_message,
    created_at,
    updated_at;
