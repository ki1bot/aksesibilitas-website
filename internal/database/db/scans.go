package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

func (queries *Queries) CreateScan(
	ctx context.Context,
	params CreateScanParams,
) (Scan, error) {
	row := queries.db.QueryRow(
		ctx,
		`
			INSERT INTO scans (
				id,
				project_id,
				created_by,
				url
			)
			VALUES ($1, $2, $3, $4)
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
				updated_at
		`,
		params.ID,
		params.ProjectID,
		params.CreatedBy,
		params.URL,
	)

	return scanScan(row)
}

func (queries *Queries) GetScanByID(
	ctx context.Context,
	id uuid.UUID,
) (Scan, error) {
	row := queries.db.QueryRow(
		ctx,
		`
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
			WHERE id = $1
			LIMIT 1
		`,
		id,
	)

	return scanScan(row)
}

func (queries *Queries) GetScanForUser(
	ctx context.Context,
	params ScanUserParams,
) (Scan, error) {
	row := queries.db.QueryRow(
		ctx,
		`
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
			WHERE scans.id = $1
			  AND project_members.user_id = $2
			LIMIT 1
		`,
		params.ScanID,
		params.UserID,
	)

	return scanScan(row)
}

func (queries *Queries) ListScansByUser(
	ctx context.Context,
	userID uuid.UUID,
	resultLimit int32,
) ([]Scan, error) {
	if resultLimit < 1 || resultLimit > 100 {
		resultLimit = 50
	}

	rows, err := queries.db.Query(
		ctx,
		`
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
			WHERE project_members.user_id = $1
			ORDER BY scans.created_at DESC
			LIMIT $2
		`,
		userID,
		resultLimit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return collectScans(rows)
}

func (queries *Queries) ListScansByProject(
	ctx context.Context,
	params ListScansByProjectParams,
) ([]Scan, error) {
	if params.ResultLimit < 1 ||
		params.ResultLimit > 100 {
		params.ResultLimit = 50
	}

	rows, err := queries.db.Query(
		ctx,
		`
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
			WHERE scans.project_id = $1
			  AND project_members.user_id = $2
			ORDER BY scans.created_at DESC
			LIMIT $3
		`,
		params.ProjectID,
		params.UserID,
		params.ResultLimit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return collectScans(rows)
}

func (queries *Queries) CountRecentScansByUser(
	ctx context.Context,
	params CountRecentScansByUserParams,
) (int64, error) {
	var count int64

	err := queries.db.QueryRow(
		ctx,
		`
			SELECT COUNT(*)
			FROM scans
			WHERE created_by = $1
			  AND created_at >= $2
		`,
		params.UserID,
		params.SinceTime,
	).Scan(&count)

	return count, err
}

func (queries *Queries) MarkScanRunning(
	ctx context.Context,
	id uuid.UUID,
) (Scan, error) {
	row := queries.db.QueryRow(
		ctx,
		`
			UPDATE scans
			SET
				status = 'running',
				error_message = '',
				started_at = NOW(),
				completed_at = NULL,
				updated_at = NOW()
			WHERE id = $1
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
				updated_at
		`,
		id,
	)

	return scanScan(row)
}

func (queries *Queries) CompleteScan(
	ctx context.Context,
	params CompleteScanParams,
) (Scan, error) {
	row := queries.db.QueryRow(
		ctx,
		`
			UPDATE scans
			SET
				status = 'completed',
				page_title = $2,
				automated_score = $3,
				critical_count = $4,
				serious_count = $5,
				moderate_count = $6,
				minor_count = $7,
				duration_ms = $8,
				error_message = '',
				completed_at = NOW(),
				updated_at = NOW()
			WHERE id = $1
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
				updated_at
		`,
		params.ID,
		params.PageTitle,
		params.AutomatedScore,
		params.CriticalCount,
		params.SeriousCount,
		params.ModerateCount,
		params.MinorCount,
		params.DurationMS,
	)

	return scanScan(row)
}

func (queries *Queries) FailScan(
	ctx context.Context,
	id uuid.UUID,
	message string,
) (Scan, error) {
	row := queries.db.QueryRow(
		ctx,
		`
			UPDATE scans
			SET
				status = 'failed',
				error_message = $2,
				completed_at = NOW(),
				updated_at = NOW()
			WHERE id = $1
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
				updated_at
		`,
		id,
		message,
	)

	return scanScan(row)
}

func (queries *Queries) CancelScan(
	ctx context.Context,
	params ScanUserParams,
) (Scan, error) {
	row := queries.db.QueryRow(
		ctx,
		`
			UPDATE scans
			SET
				status = 'cancelled',
				error_message = '',
				completed_at = NOW(),
				updated_at = NOW()
			WHERE id = $1
			  AND status IN ('queued', 'running')
			  AND EXISTS (
				  SELECT 1
				  FROM project_members
				  WHERE project_members.project_id = scans.project_id
				    AND project_members.user_id = $2
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
				updated_at
		`,
		params.ScanID,
		params.UserID,
	)

	return scanScan(row)
}

func (queries *Queries) ResetScanForRetry(
	ctx context.Context,
	params ScanUserParams,
) (Scan, error) {
	row := queries.db.QueryRow(
		ctx,
		`
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
			WHERE id = $1
			  AND status IN ('failed', 'cancelled')
			  AND EXISTS (
				  SELECT 1
				  FROM project_members
				  WHERE project_members.project_id = scans.project_id
				    AND project_members.user_id = $2
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
				updated_at
		`,
		params.ScanID,
		params.UserID,
	)

	return scanScan(row)
}

func (queries *Queries) DeleteScanForUser(
	ctx context.Context,
	params ScanUserParams,
) error {
	tag, err := queries.db.Exec(
		ctx,
		`
			DELETE FROM scans
			WHERE id = $1
			  AND EXISTS (
				  SELECT 1
				  FROM project_members
				  WHERE project_members.project_id = scans.project_id
				    AND project_members.user_id = $2
			  )
		`,
		params.ScanID,
		params.UserID,
	)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func collectScans(rows interface {
	Next() bool
	Scan(...any) error
	Err() error
}) ([]Scan, error) {
	scans := make([]Scan, 0)

	for rows.Next() {
		scan, err := scanScan(rows)
		if err != nil {
			return nil, err
		}

		scans = append(scans, scan)
	}

	return scans, rows.Err()
}

func scanScan(row interface {
	Scan(...any) error
}) (Scan, error) {
	var scan Scan

	err := row.Scan(
		&scan.ID,
		&scan.ProjectID,
		&scan.CreatedBy,
		&scan.URL,
		&scan.Status,
		&scan.PageTitle,
		&scan.AutomatedScore,
		&scan.CriticalCount,
		&scan.SeriousCount,
		&scan.ModerateCount,
		&scan.MinorCount,
		&scan.ErrorMessage,
		&scan.DurationMS,
		&scan.CreatedAt,
		&scan.UpdatedAt,
	)

	return scan, err
}

func NewRecentScanWindow() time.Time {
	return time.Now().Add(-time.Minute)
}
