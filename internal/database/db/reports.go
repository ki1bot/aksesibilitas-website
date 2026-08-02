package db

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

func (queries *Queries) CreateReport(
	ctx context.Context,
	params CreateReportParams,
) (Report, error) {
	var report Report

	err := queries.db.QueryRow(
		ctx,
		`
			INSERT INTO reports (
				id,
				scan_id,
				created_by,
				format,
				filename,
				content_type,
				content
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING
				id,
				scan_id,
				created_by,
				format,
				filename,
				content_type,
				content,
				created_at
		`,
		params.ID,
		params.ScanID,
		params.CreatedBy,
		params.Format,
		params.Filename,
		params.ContentType,
		params.Content,
	).Scan(
		&report.ID,
		&report.ScanID,
		&report.CreatedBy,
		&report.Format,
		&report.Filename,
		&report.ContentType,
		&report.Content,
		&report.CreatedAt,
	)

	return report, err
}

func (queries *Queries) GetReportForUser(
	ctx context.Context,
	params ReportUserParams,
) (Report, error) {
	var report Report

	err := queries.db.QueryRow(
		ctx,
		`
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
			WHERE reports.id = $1
			  AND project_members.user_id = $2
			LIMIT 1
		`,
		params.ReportID,
		params.UserID,
	).Scan(
		&report.ID,
		&report.ScanID,
		&report.CreatedBy,
		&report.Format,
		&report.Filename,
		&report.ContentType,
		&report.Content,
		&report.CreatedAt,
	)

	return report, err
}

func (queries *Queries) CreateActivityLog(
	ctx context.Context,
	params CreateActivityLogParams,
) error {
	metadata := params.Metadata
	if len(metadata) == 0 {
		metadata = json.RawMessage(`{}`)
	}

	_, err := queries.db.Exec(
		ctx,
		`
			INSERT INTO activity_logs (
				id,
				user_id,
				project_id,
				action,
				entity_type,
				entity_id,
				metadata
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
		params.ID,
		params.UserID,
		params.ProjectID,
		params.Action,
		params.EntityType,
		params.EntityID,
		metadata,
	)

	return err
}

func NewActivityID() uuid.UUID {
	return uuid.New()
}
