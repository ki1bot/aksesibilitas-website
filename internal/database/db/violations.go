package db

import (
	"context"

	"github.com/google/uuid"
)

func (queries *Queries) CreateViolation(
	ctx context.Context,
	params CreateViolationParams,
) (Violation, error) {
	row := queries.db.QueryRow(
		ctx,
		`
			INSERT INTO violations (
				id,
				scanned_page_id,
				rule_id,
				impact,
				description,
				help,
				help_url,
				tags
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
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
				updated_at
		`,
		params.ID,
		params.ScannedPageID,
		params.RuleID,
		params.Impact,
		params.Description,
		params.Help,
		params.HelpURL,
		params.Tags,
	)

	return scanViolation(row)
}

func (queries *Queries) CreateViolationNode(
	ctx context.Context,
	params CreateViolationNodeParams,
) (ViolationNode, error) {
	var node ViolationNode

	err := queries.db.QueryRow(
		ctx,
		`
			INSERT INTO violation_nodes (
				id,
				violation_id,
				html,
				target,
				failure_summary
			)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING
				id,
				violation_id,
				html,
				target,
				failure_summary,
				created_at
		`,
		params.ID,
		params.ViolationID,
		params.HTML,
		params.Target,
		params.FailureSummary,
	).Scan(
		&node.ID,
		&node.ViolationID,
		&node.HTML,
		&node.Target,
		&node.FailureSummary,
		&node.CreatedAt,
	)

	return node, err
}

func (queries *Queries) ListViolationsForScan(
	ctx context.Context,
	params ScanUserParams,
) ([]Violation, error) {
	rows, err := queries.db.Query(
		ctx,
		`
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
			WHERE scans.id = $1
			  AND project_members.user_id = $2
			ORDER BY
				CASE violations.impact
					WHEN 'critical' THEN 1
					WHEN 'serious' THEN 2
					WHEN 'moderate' THEN 3
					WHEN 'minor' THEN 4
				END,
				violations.rule_id
		`,
		params.ScanID,
		params.UserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	violations := make([]Violation, 0)

	for rows.Next() {
		violation, scanErr := scanViolation(rows)
		if scanErr != nil {
			return nil, scanErr
		}

		violations = append(violations, violation)
	}

	return violations, rows.Err()
}

func (queries *Queries) GetViolationForUser(
	ctx context.Context,
	params ViolationUserParams,
) (Violation, error) {
	row := queries.db.QueryRow(
		ctx,
		`
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
			WHERE violations.id = $1
			  AND project_members.user_id = $2
			LIMIT 1
		`,
		params.ViolationID,
		params.UserID,
	)

	return scanViolation(row)
}

func (queries *Queries) ListViolationNodes(
	ctx context.Context,
	violationID uuid.UUID,
) ([]ViolationNode, error) {
	rows, err := queries.db.Query(
		ctx,
		`
			SELECT
				id,
				violation_id,
				html,
				target,
				failure_summary,
				created_at
			FROM violation_nodes
			WHERE violation_id = $1
			ORDER BY created_at
		`,
		violationID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodes := make([]ViolationNode, 0)

	for rows.Next() {
		var node ViolationNode

		if err := rows.Scan(
			&node.ID,
			&node.ViolationID,
			&node.HTML,
			&node.Target,
			&node.FailureSummary,
			&node.CreatedAt,
		); err != nil {
			return nil, err
		}

		nodes = append(nodes, node)
	}

	return nodes, rows.Err()
}

func (queries *Queries) UpdateViolationReview(
	ctx context.Context,
	params UpdateViolationReviewParams,
) (Violation, error) {
	row := queries.db.QueryRow(
		ctx,
		`
			UPDATE violations
			SET
				review_status = $3,
				notes = $4,
				updated_at = NOW()
			WHERE id = $1
			  AND EXISTS (
				  SELECT 1
				  FROM scanned_pages
				  INNER JOIN scans
					  ON scans.id = scanned_pages.scan_id
				  INNER JOIN project_members
					  ON project_members.project_id = scans.project_id
				  WHERE scanned_pages.id = violations.scanned_page_id
				    AND project_members.user_id = $2
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
				updated_at
		`,
		params.ViolationID,
		params.UserID,
		params.ReviewStatus,
		params.Notes,
	)

	return scanViolation(row)
}

func scanViolation(row interface {
	Scan(...any) error
}) (Violation, error) {
	var violation Violation

	err := row.Scan(
		&violation.ID,
		&violation.ScannedPageID,
		&violation.RuleID,
		&violation.Impact,
		&violation.Description,
		&violation.Help,
		&violation.HelpURL,
		&violation.Tags,
		&violation.ReviewStatus,
		&violation.Notes,
		&violation.CreatedAt,
		&violation.UpdatedAt,
	)

	return violation, err
}
