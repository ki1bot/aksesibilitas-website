package db

import (
	"context"

	"github.com/google/uuid"
)

func (queries *Queries) DeleteScanResults(
	ctx context.Context,
	scanID uuid.UUID,
) error {
	_, err := queries.db.Exec(
		ctx,
		`
			DELETE FROM scanned_pages
			WHERE scan_id = $1
		`,
		scanID,
	)

	return err
}

func (queries *Queries) DeleteScanReports(
	ctx context.Context,
	scanID uuid.UUID,
) error {
	_, err := queries.db.Exec(
		ctx,
		`
			DELETE FROM reports
			WHERE scan_id = $1
		`,
		scanID,
	)

	return err
}

func (queries *Queries) CreateScannedPage(
	ctx context.Context,
	params CreateScannedPageParams,
) (ScannedPage, error) {
	var page ScannedPage

	err := queries.db.QueryRow(
		ctx,
		`
			INSERT INTO scanned_pages (
				id,
				scan_id,
				url,
				title,
				language
			)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING
				id,
				scan_id,
				url,
				title,
				language,
				created_at
		`,
		params.ID,
		params.ScanID,
		params.URL,
		params.Title,
		params.Language,
	).Scan(
		&page.ID,
		&page.ScanID,
		&page.URL,
		&page.Title,
		&page.Language,
		&page.CreatedAt,
	)

	return page, err
}
