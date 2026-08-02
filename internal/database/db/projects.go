package db

import (
	"context"

	"github.com/google/uuid"
)

func (queries *Queries) CreateProject(
	ctx context.Context,
	params CreateProjectParams,
) (Project, error) {
	row := queries.db.QueryRow(
		ctx,
		`
			INSERT INTO projects (
				id,
				owner_id,
				name,
				description
			)
			VALUES ($1, $2, $3, $4)
			RETURNING
				id,
				owner_id,
				name,
				description,
				created_at,
				updated_at
		`,
		params.ID,
		params.OwnerID,
		params.Name,
		params.Description,
	)

	return scanProject(row)
}

func (queries *Queries) AddProjectMember(
	ctx context.Context,
	params AddProjectMemberParams,
) error {
	_, err := queries.db.Exec(
		ctx,
		`
			INSERT INTO project_members (
				project_id,
				user_id,
				role
			)
			VALUES ($1, $2, $3)
			ON CONFLICT (project_id, user_id)
			DO UPDATE SET role = EXCLUDED.role
		`,
		params.ProjectID,
		params.UserID,
		params.Role,
	)

	return err
}

func (queries *Queries) ListProjectsByUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]Project, error) {
	rows, err := queries.db.Query(
		ctx,
		`
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
			WHERE project_members.user_id = $1
			ORDER BY projects.updated_at DESC
		`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := make([]Project, 0)

	for rows.Next() {
		project, scanErr := scanProject(rows)
		if scanErr != nil {
			return nil, scanErr
		}

		projects = append(projects, project)
	}

	return projects, rows.Err()
}

func (queries *Queries) GetProjectForUser(
	ctx context.Context,
	params ProjectUserParams,
) (Project, error) {
	row := queries.db.QueryRow(
		ctx,
		`
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
			WHERE projects.id = $1
			  AND project_members.user_id = $2
			LIMIT 1
		`,
		params.ProjectID,
		params.UserID,
	)

	return scanProject(row)
}

func (queries *Queries) GetOwnedProject(
	ctx context.Context,
	params ProjectUserParams,
) (Project, error) {
	row := queries.db.QueryRow(
		ctx,
		`
			SELECT
				id,
				owner_id,
				name,
				description,
				created_at,
				updated_at
			FROM projects
			WHERE id = $1
			  AND owner_id = $2
			LIMIT 1
		`,
		params.ProjectID,
		params.UserID,
	)

	return scanProject(row)
}

func (queries *Queries) UpdateProject(
	ctx context.Context,
	params UpdateProjectParams,
) (Project, error) {
	row := queries.db.QueryRow(
		ctx,
		`
			UPDATE projects
			SET
				name = $3,
				description = $4,
				updated_at = NOW()
			WHERE id = $1
			  AND owner_id = $2
			RETURNING
				id,
				owner_id,
				name,
				description,
				created_at,
				updated_at
		`,
		params.ProjectID,
		params.UserID,
		params.Name,
		params.Description,
	)

	return scanProject(row)
}

func (queries *Queries) DeleteProject(
	ctx context.Context,
	params ProjectUserParams,
) error {
	tag, err := queries.db.Exec(
		ctx,
		`
			DELETE FROM projects
			WHERE id = $1
			  AND owner_id = $2
		`,
		params.ProjectID,
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

func scanProject(row interface {
	Scan(...any) error
}) (Project, error) {
	var project Project

	err := row.Scan(
		&project.ID,
		&project.OwnerID,
		&project.Name,
		&project.Description,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	return project, err
}
