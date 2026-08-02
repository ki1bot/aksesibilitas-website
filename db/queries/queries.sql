-- name: CreateUser :one
INSERT INTO users (
    id,
    name,
    email,
    password_hash
) VALUES (
    sqlc.arg(id),
    sqlc.arg(name),
    LOWER(sqlc.arg(email)),
    sqlc.arg(password_hash)
)
RETURNING
    id,
    name,
    email,
    password_hash,
    created_at,
    updated_at;

-- name: GetUserByID :one
SELECT
    id,
    name,
    email,
    password_hash,
    created_at,
    updated_at
FROM users
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: GetUserByEmail :one
SELECT
    id,
    name,
    email,
    password_hash,
    created_at,
    updated_at
FROM users
WHERE LOWER(email) = LOWER(sqlc.arg(email))
LIMIT 1;

-- name: CreateSession :one
INSERT INTO sessions (
    id,
    user_id,
    token_hash,
    csrf_hash,
    user_agent,
    ip_address,
    expires_at
) VALUES (
    sqlc.arg(id),
    sqlc.arg(user_id),
    sqlc.arg(token_hash),
    sqlc.arg(csrf_hash),
    sqlc.arg(user_agent),
    sqlc.arg(ip_address),
    sqlc.arg(expires_at)
)
RETURNING
    id,
    user_id,
    token_hash,
    csrf_hash,
    user_agent,
    ip_address,
    expires_at,
    last_used_at,
    created_at;

-- name: GetSessionByTokenHash :one
SELECT
    id,
    user_id,
    token_hash,
    csrf_hash,
    user_agent,
    ip_address,
    expires_at,
    last_used_at,
    created_at
FROM sessions
WHERE token_hash = sqlc.arg(token_hash)
  AND expires_at > NOW()
LIMIT 1;

-- name: TouchSession :exec
UPDATE sessions
SET last_used_at = NOW()
WHERE id = sqlc.arg(id);

-- name: DeleteSessionByTokenHash :exec
DELETE FROM sessions
WHERE token_hash = sqlc.arg(token_hash);

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at <= NOW();

-- name: CreateProject :one
INSERT INTO projects (
    id,
    owner_id,
    name,
    description
) VALUES (
    sqlc.arg(id),
    sqlc.arg(owner_id),
    sqlc.arg(name),
    sqlc.arg(description)
)
RETURNING
    id,
    owner_id,
    name,
    description,
    created_at,
    updated_at;

-- name: AddProjectMember :exec
INSERT INTO project_members (
    project_id,
    user_id,
    role
) VALUES (
    sqlc.arg(project_id),
    sqlc.arg(user_id),
    sqlc.arg(role)
)
ON CONFLICT (project_id, user_id)
DO UPDATE SET role = EXCLUDED.role;

-- name: ListProjectsByUser :many
SELECT DISTINCT
    projects.id,
    projects.owner_id,
    projects.name,
    projects.description,
    projects.created_at,
    projects.updated_at
FROM projects
INNER JOIN project_members
    ON project_members.project_id = projects.id
WHERE project_members.user_id = sqlc.arg(user_id)
ORDER BY projects.updated_at DESC;

-- name: GetProjectForUser :one
SELECT
    projects.id,
    projects.owner_id,
    projects.name,
    projects.description,
    projects.created_at,
    projects.updated_at
FROM projects
INNER JOIN project_members
    ON project_members.project_id = projects.id
WHERE projects.id = sqlc.arg(project_id)
  AND project_members.user_id = sqlc.arg(user_id)
LIMIT 1;

-- name: GetOwnedProject :one
SELECT
    id,
    owner_id,
    name,
    description,
    created_at,
    updated_at
FROM projects
WHERE id = sqlc.arg(project_id)
  AND owner_id = sqlc.arg(user_id)
LIMIT 1;

-- name: UpdateProject :one
UPDATE projects
SET
    name = sqlc.arg(name),
    description = sqlc.arg(description),
    updated_at = NOW()
WHERE id = sqlc.arg(project_id)
  AND owner_id = sqlc.arg(user_id)
RETURNING
    id,
    owner_id,
    name,
    description,
    created_at,
    updated_at;

-- name: DeleteProject :exec
DELETE FROM projects
WHERE id = sqlc.arg(project_id)
  AND owner_id = sqlc.arg(user_id);

-- name: CreateScan :one
INSERT INTO scans (
    id,
    project_id,
    created_by,
    url
) VALUES (
    sqlc.arg(id),
    sqlc.arg(project_id),
    sqlc.arg(created_by),
    sqlc.arg(url)
)
RETURNING
    id,
    project_id,
    created_by,
    url,
    status,
    page_title,
    automated_score,
    critical_count,
    serious_count,
    moderate_count,
    minor_count,
    error_message,
    duration_ms,
    created_at,
    updated_at;

-- name: GetScanByID :one
SELECT
    id,
    project_id,
    created_by,
    url,
    status,
    page_title,
    automated_score,
    critical_count,
    serious_count,
    moderate_count,
    minor_count,
    error_message,
    duration_ms,
    created_at,
    updated_at
FROM scans
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: GetScanForUser :one
SELECT
    scans.id,
    scans.project_id,
    scans.created_by,
    scans.url,
    scans.status,
    scans.page_title,
    scans.automated_score,
    scans.critical_count,
    scans.serious_count,
    scans.moderate_count,
    scans.minor_count,
    scans.error_message,
    scans.duration_ms,
    scans.created_at,
    scans.updated_at
FROM scans
INNER JOIN project_members
    ON project_members.project_id = scans.project_id
WHERE scans.id = sqlc.arg(scan_id)
  AND project_members.user_id = sqlc.arg(user_id)
LIMIT 1;

-- name: ListScansByUser :many
SELECT
    scans.id,
    scans.project_id,
    scans.created_by,
    scans.url,
    scans.status,
    scans.page_title,
    scans.automated_score,
    scans.critical_count,
    scans.serious_count,
    scans.moderate_count,
    scans.minor_count,
    scans.error_message,
    scans.duration_ms,
    scans.created_at,
    scans.updated_at
FROM scans
INNER JOIN project_members
    ON project_members.project_id = scans.project_id
WHERE project_members.user_id = sqlc.arg(user_id)
ORDER BY scans.created_at DESC
LIMIT sqlc.arg(result_limit);

-- name: ListScansByProject :many
SELECT
    scans.id,
    scans.project_id,
    scans.created_by,
    scans.url,
    scans.status,
    scans.page_title,
    scans.automated_score,
    scans.critical_count,
    scans.serious_count,
    scans.moderate_count,
    scans.minor_count,
    scans.error_message,
    scans.duration_ms,
    scans.created_at,
    scans.updated_at
FROM scans
INNER JOIN project_members
    ON project_members.project_id = scans.project_id
WHERE scans.project_id = sqlc.arg(project_id)
  AND project_members.user_id = sqlc.arg(user_id)
ORDER BY scans.created_at DESC
LIMIT sqlc.arg(result_limit);

-- name: CountRecentScansByUser :one
SELECT COUNT(*)
FROM scans
WHERE created_by = sqlc.arg(user_id)
  AND created_at >= sqlc.arg(since_time);

-- name: MarkScanRunning :one
UPDATE scans
SET
    status = 'running',
    error_message = '',
    started_at = NOW(),
    completed_at = NULL,
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND status IN ('queued', 'failed')
RETURNING
    id,
    project_id,
    created_by,
    url,
    status,
    page_title,
    automated_score,
    critical_count,
    serious_count,
    moderate_count,
    minor_count,
    error_message,
    duration_ms,
    created_at,
    updated_at;

-- name: CompleteScan :one
UPDATE scans
SET
    status = 'completed',
    page_title = sqlc.arg(page_title),
    automated_score = sqlc.arg(automated_score),
    critical_count = sqlc.arg(critical_count),
    serious_count = sqlc.arg(serious_count),
    moderate_count = sqlc.arg(moderate_count),
    minor_count = sqlc.arg(minor_count),
    duration_ms = sqlc.arg(duration_ms),
    error_message = '',
    completed_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND status = 'running'
RETURNING
    id,
    project_id,
    created_by,
    url,
    status,
    page_title,
    automated_score,
    critical_count,
    serious_count,
    moderate_count,
    minor_count,
    error_message,
    duration_ms,
    created_at,
    updated_at;

-- name: FailScan :one
UPDATE scans
SET
    status = 'failed',
    error_message = sqlc.arg(error_message),
    completed_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND status <> 'cancelled'
RETURNING
    id,
    project_id,
    created_by,
    url,
    status,
    page_title,
    automated_score,
    critical_count,
    serious_count,
    moderate_count,
    minor_count,
    error_message,
    duration_ms,
    created_at,
    updated_at;

-- name: CancelScan :one
UPDATE scans
SET
    status = 'cancelled',
    error_message = '',
    completed_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg(scan_id)
  AND status IN ('queued', 'running')
  AND EXISTS (
      SELECT 1
      FROM project_members
      WHERE project_members.project_id = scans.project_id
        AND project_members.user_id = sqlc.arg(user_id)
  )
RETURNING
    id,
    project_id,
    created_by,
    url,
    status,
    page_title,
    automated_score,
    critical_count,
    serious_count,
    moderate_count,
    minor_count,
    error_message,
    duration_ms,
    created_at,
    updated_at;

-- name: ResetScanForRetry :one
UPDATE scans
SET
    status = 'queued',
    page_title = '',
    automated_score = 0,
    critical_count = 0,
    serious_count = 0,
    moderate_count = 0,
    minor_count = 0,
    duration_ms = 0,
    error_message = '',
    started_at = NULL,
    completed_at = NULL,
    updated_at = NOW()
WHERE id = sqlc.arg(scan_id)
  AND status IN ('failed', 'cancelled')
  AND EXISTS (
      SELECT 1
      FROM project_members
      WHERE project_members.project_id = scans.project_id
        AND project_members.user_id = sqlc.arg(user_id)
  )
RETURNING
    id,
    project_id,
    created_by,
    url,
    status,
    page_title,
    automated_score,
    critical_count,
    serious_count,
    moderate_count,
    minor_count,
    error_message,
    duration_ms,
    created_at,
    updated_at;

-- name: DeleteScanForUser :exec
DELETE FROM scans
WHERE id = sqlc.arg(scan_id)
  AND EXISTS (
      SELECT 1
      FROM project_members
      WHERE project_members.project_id = scans.project_id
        AND project_members.user_id = sqlc.arg(user_id)
  );

-- name: DeleteScanResults :exec
DELETE FROM scanned_pages
WHERE scan_id = sqlc.arg(scan_id);

-- name: DeleteScanReports :exec
DELETE FROM reports
WHERE scan_id = sqlc.arg(scan_id);

-- name: CreateScannedPage :one
INSERT INTO scanned_pages (
    id,
    scan_id,
    url,
    title,
    language
) VALUES (
    sqlc.arg(id),
    sqlc.arg(scan_id),
    sqlc.arg(url),
    sqlc.arg(title),
    sqlc.arg(language)
)
RETURNING
    id,
    scan_id,
    url,
    title,
    language,
    created_at;

-- name: CreateViolation :one
INSERT INTO violations (
    id,
    scanned_page_id,
    rule_id,
    impact,
    description,
    help,
    help_url,
    tags
) VALUES (
    sqlc.arg(id),
    sqlc.arg(scanned_page_id),
    sqlc.arg(rule_id),
    sqlc.arg(impact),
    sqlc.arg(description),
    sqlc.arg(help),
    sqlc.arg(help_url),
    sqlc.arg(tags)
)
RETURNING
    id,
    scanned_page_id,
    rule_id,
    impact,
    description,
    help,
    help_url,
    tags,
    review_status,
    notes,
    created_at,
    updated_at;

-- name: CreateViolationNode :one
INSERT INTO violation_nodes (
    id,
    violation_id,
    html,
    target,
    failure_summary
) VALUES (
    sqlc.arg(id),
    sqlc.arg(violation_id),
    sqlc.arg(html),
    sqlc.arg(target),
    sqlc.arg(failure_summary)
)
RETURNING
    id,
    violation_id,
    html,
    target,
    failure_summary,
    created_at;

-- name: ListViolationsForScan :many
SELECT
    violations.id,
    violations.scanned_page_id,
    violations.rule_id,
    violations.impact,
    violations.description,
    violations.help,
    violations.help_url,
    violations.tags,
    violations.review_status,
    violations.notes,
    violations.created_at,
    violations.updated_at
FROM violations
INNER JOIN scanned_pages
    ON scanned_pages.id = violations.scanned_page_id
INNER JOIN scans
    ON scans.id = scanned_pages.scan_id
INNER JOIN project_members
    ON project_members.project_id = scans.project_id
WHERE scans.id = sqlc.arg(scan_id)
  AND project_members.user_id = sqlc.arg(user_id)
ORDER BY
    CASE violations.impact
        WHEN 'critical' THEN 1
        WHEN 'serious' THEN 2
        WHEN 'moderate' THEN 3
        WHEN 'minor' THEN 4
    END,
    violations.rule_id;

-- name: GetViolationForUser :one
SELECT
    violations.id,
    violations.scanned_page_id,
    violations.rule_id,
    violations.impact,
    violations.description,
    violations.help,
    violations.help_url,
    violations.tags,
    violations.review_status,
    violations.notes,
    violations.created_at,
    violations.updated_at
FROM violations
INNER JOIN scanned_pages
    ON scanned_pages.id = violations.scanned_page_id
INNER JOIN scans
    ON scans.id = scanned_pages.scan_id
INNER JOIN project_members
    ON project_members.project_id = scans.project_id
WHERE violations.id = sqlc.arg(violation_id)
  AND project_members.user_id = sqlc.arg(user_id)
LIMIT 1;

-- name: ListViolationNodes :many
SELECT
    id,
    violation_id,
    html,
    target,
    failure_summary,
    created_at
FROM violation_nodes
WHERE violation_id = sqlc.arg(violation_id)
ORDER BY created_at;

-- name: UpdateViolationReview :one
UPDATE violations
SET
    review_status = sqlc.arg(review_status),
    notes = sqlc.arg(notes),
    updated_at = NOW()
WHERE violations.id = sqlc.arg(violation_id)
  AND EXISTS (
      SELECT 1
      FROM scanned_pages
      INNER JOIN scans
          ON scans.id = scanned_pages.scan_id
      INNER JOIN project_members
          ON project_members.project_id = scans.project_id
      WHERE scanned_pages.id = violations.scanned_page_id
        AND project_members.user_id = sqlc.arg(user_id)
  )
RETURNING
    violations.id,
    violations.scanned_page_id,
    violations.rule_id,
    violations.impact,
    violations.description,
    violations.help,
    violations.help_url,
    violations.tags,
    violations.review_status,
    violations.notes,
    violations.created_at,
    violations.updated_at;

-- name: CreateManualReview :one
INSERT INTO manual_reviews (
    id,
    scan_id
) VALUES (
    sqlc.arg(id),
    sqlc.arg(scan_id)
)
RETURNING
    id,
    scan_id,
    status,
    created_at,
    updated_at;

-- name: CreateManualReviewItem :one
INSERT INTO manual_review_items (
    id,
    manual_review_id,
    criterion,
    instruction,
    position
) VALUES (
    sqlc.arg(id),
    sqlc.arg(manual_review_id),
    sqlc.arg(criterion),
    sqlc.arg(instruction),
    sqlc.arg(position)
)
RETURNING
    id,
    manual_review_id,
    criterion,
    instruction,
    status,
    notes,
    position,
    created_at,
    updated_at;

-- name: GetManualReviewForScan :one
SELECT
    manual_reviews.id,
    manual_reviews.scan_id,
    manual_reviews.status,
    manual_reviews.created_at,
    manual_reviews.updated_at
FROM manual_reviews
INNER JOIN scans
    ON scans.id = manual_reviews.scan_id
INNER JOIN project_members
    ON project_members.project_id = scans.project_id
WHERE manual_reviews.scan_id = sqlc.arg(scan_id)
  AND project_members.user_id = sqlc.arg(user_id)
LIMIT 1;

-- name: ListManualReviewItems :many
SELECT
    id,
    manual_review_id,
    criterion,
    instruction,
    status,
    notes,
    position,
    created_at,
    updated_at
FROM manual_review_items
WHERE manual_review_id = sqlc.arg(manual_review_id)
ORDER BY position;

-- name: UpdateManualReviewItem :one
UPDATE manual_review_items
SET
    status = sqlc.arg(status),
    notes = sqlc.arg(notes),
    updated_at = NOW()
WHERE manual_review_items.id = sqlc.arg(item_id)
  AND EXISTS (
      SELECT 1
      FROM manual_reviews
      INNER JOIN scans
          ON scans.id = manual_reviews.scan_id
      INNER JOIN project_members
          ON project_members.project_id = scans.project_id
      WHERE manual_reviews.id = manual_review_items.manual_review_id
        AND project_members.user_id = sqlc.arg(user_id)
  )
RETURNING
    manual_review_items.id,
    manual_review_items.manual_review_id,
    manual_review_items.criterion,
    manual_review_items.instruction,
    manual_review_items.status,
    manual_review_items.notes,
    manual_review_items.position,
    manual_review_items.created_at,
    manual_review_items.updated_at;

-- name: RefreshManualReviewStatus :exec
UPDATE manual_reviews
SET
    status = CASE
        WHEN EXISTS (
            SELECT 1
            FROM manual_review_items
            WHERE manual_review_items.manual_review_id =
                sqlc.arg(target_manual_review_id)
              AND manual_review_items.status = 'failed'
        ) THEN 'failed'::review_status
        WHEN EXISTS (
            SELECT 1
            FROM manual_review_items
            WHERE manual_review_items.manual_review_id =
                sqlc.arg(target_manual_review_id)
              AND manual_review_items.status = 'pending'
        ) THEN 'pending'::review_status
        ELSE 'passed'::review_status
    END,
    updated_at = NOW()
WHERE manual_reviews.id = sqlc.arg(target_manual_review_id);

-- name: ResetManualReview :exec
UPDATE manual_review_items
SET
    status = 'pending',
    notes = '',
    updated_at = NOW()
WHERE manual_review_items.manual_review_id = (
    SELECT manual_reviews.id
    FROM manual_reviews
    WHERE manual_reviews.scan_id = sqlc.arg(scan_id)
);

-- name: ResetManualReviewStatus :exec
UPDATE manual_reviews
SET
    status = 'pending',
    updated_at = NOW()
WHERE scan_id = sqlc.arg(scan_id);

-- name: CreateReport :one
INSERT INTO reports (
    id,
    scan_id,
    created_by,
    format,
    filename,
    content_type,
    content
) VALUES (
    sqlc.arg(id),
    sqlc.arg(scan_id),
    sqlc.arg(created_by),
    sqlc.arg(format),
    sqlc.arg(filename),
    sqlc.arg(content_type),
    sqlc.arg(content)
)
RETURNING
    id,
    scan_id,
    created_by,
    format,
    filename,
    content_type,
    content,
    created_at;

-- name: GetReportForUser :one
SELECT
    reports.id,
    reports.scan_id,
    reports.created_by,
    reports.format,
    reports.filename,
    reports.content_type,
    reports.content,
    reports.created_at
FROM reports
INNER JOIN scans
    ON scans.id = reports.scan_id
INNER JOIN project_members
    ON project_members.project_id = scans.project_id
WHERE reports.id = sqlc.arg(report_id)
  AND project_members.user_id = sqlc.arg(user_id)
LIMIT 1;

-- name: CreateActivityLog :exec
INSERT INTO activity_logs (
    id,
    user_id,
    project_id,
    action,
    entity_type,
    entity_id,
    metadata
) VALUES (
    sqlc.arg(id),
    sqlc.arg(user_id),
    sqlc.arg(project_id),
    sqlc.arg(action),
    sqlc.arg(entity_type),
    sqlc.arg(entity_id),
    sqlc.arg(metadata)
);