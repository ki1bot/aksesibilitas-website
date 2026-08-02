package db

import (
	"context"

	"github.com/google/uuid"
)

func (queries *Queries) CreateManualReview(
	ctx context.Context,
	params CreateManualReviewParams,
) (ManualReview, error) {
	var review ManualReview

	err := queries.db.QueryRow(
		ctx,
		`
			INSERT INTO manual_reviews (
				id,
				scan_id
			)
			VALUES ($1, $2)
			RETURNING
				id,
				scan_id,
				status,
				created_at,
				updated_at
		`,
		params.ID,
		params.ScanID,
	).Scan(
		&review.ID,
		&review.ScanID,
		&review.Status,
		&review.CreatedAt,
		&review.UpdatedAt,
	)

	return review, err
}

func (queries *Queries) CreateManualReviewItem(
	ctx context.Context,
	params CreateManualReviewItemParams,
) (ManualReviewItem, error) {
	var item ManualReviewItem

	err := queries.db.QueryRow(
		ctx,
		`
			INSERT INTO manual_review_items (
				id,
				manual_review_id,
				criterion,
				instruction,
				position
			)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING
				id,
				manual_review_id,
				criterion,
				instruction,
				status,
				notes,
				position,
				created_at,
				updated_at
		`,
		params.ID,
		params.ManualReviewID,
		params.Criterion,
		params.Instruction,
		params.Position,
	).Scan(
		&item.ID,
		&item.ManualReviewID,
		&item.Criterion,
		&item.Instruction,
		&item.Status,
		&item.Notes,
		&item.Position,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	return item, err
}

func (queries *Queries) GetManualReviewForScan(
	ctx context.Context,
	params ScanUserParams,
) (ManualReview, error) {
	var review ManualReview

	err := queries.db.QueryRow(
		ctx,
		`
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
			WHERE manual_reviews.scan_id = $1
			  AND project_members.user_id = $2
			LIMIT 1
		`,
		params.ScanID,
		params.UserID,
	).Scan(
		&review.ID,
		&review.ScanID,
		&review.Status,
		&review.CreatedAt,
		&review.UpdatedAt,
	)

	return review, err
}

func (queries *Queries) ListManualReviewItems(
	ctx context.Context,
	manualReviewID uuid.UUID,
) ([]ManualReviewItem, error) {
	rows, err := queries.db.Query(
		ctx,
		`
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
			WHERE manual_review_id = $1
			ORDER BY position
		`,
		manualReviewID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]ManualReviewItem, 0)

	for rows.Next() {
		var item ManualReviewItem

		if err := rows.Scan(
			&item.ID,
			&item.ManualReviewID,
			&item.Criterion,
			&item.Instruction,
			&item.Status,
			&item.Notes,
			&item.Position,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

func (queries *Queries) UpdateManualReviewItem(
	ctx context.Context,
	params UpdateManualReviewItemParams,
) (ManualReviewItem, error) {
	var item ManualReviewItem

	err := queries.db.QueryRow(
		ctx,
		`
			UPDATE manual_review_items
			SET
				status = $3,
				notes = $4,
				updated_at = NOW()
			WHERE id = $1
			  AND EXISTS (
				  SELECT 1
				  FROM manual_reviews
				  INNER JOIN scans
					  ON scans.id = manual_reviews.scan_id
				  INNER JOIN project_members
					  ON project_members.project_id = scans.project_id
				  WHERE manual_reviews.id = manual_review_items.manual_review_id
				    AND project_members.user_id = $2
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
				updated_at
		`,
		params.ItemID,
		params.UserID,
		params.Status,
		params.Notes,
	).Scan(
		&item.ID,
		&item.ManualReviewID,
		&item.Criterion,
		&item.Instruction,
		&item.Status,
		&item.Notes,
		&item.Position,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	return item, err
}

func (queries *Queries) RefreshManualReviewStatus(
	ctx context.Context,
	manualReviewID uuid.UUID,
) error {
	_, err := queries.db.Exec(
		ctx,
		`
			UPDATE manual_reviews
			SET
				status = CASE
					WHEN EXISTS (
						SELECT 1
						FROM manual_review_items
						WHERE manual_review_id = $1
						  AND status = 'failed'
					) THEN 'failed'::review_status
					WHEN EXISTS (
						SELECT 1
						FROM manual_review_items
						WHERE manual_review_id = $1
						  AND status = 'pending'
					) THEN 'pending'::review_status
					ELSE 'passed'::review_status
				END,
				updated_at = NOW()
			WHERE id = $1
		`,
		manualReviewID,
	)

	return err
}

func (queries *Queries) ResetManualReview(
	ctx context.Context,
	scanID uuid.UUID,
) error {
	_, err := queries.db.Exec(
		ctx,
		`
			UPDATE manual_review_items
			SET
				status = 'pending',
				notes = '',
				updated_at = NOW()
			WHERE manual_review_id = (
				SELECT id
				FROM manual_reviews
				WHERE scan_id = $1
			);

			UPDATE manual_reviews
			SET
				status = 'pending',
				updated_at = NOW()
			WHERE scan_id = $1
		`,
		scanID,
	)

	return err
}
