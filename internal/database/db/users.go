package db

import (
	"context"

	"github.com/google/uuid"
)

func (queries *Queries) CreateUser(
	ctx context.Context,
	params CreateUserParams,
) (User, error) {
	row := queries.db.QueryRow(
		ctx,
		`
			INSERT INTO users (
				id,
				name,
				email,
				password_hash
			)
			VALUES ($1, $2, LOWER($3), $4)
			RETURNING
				id,
				name,
				email,
				password_hash,
				created_at,
				updated_at
		`,
		params.ID,
		params.Name,
		params.Email,
		params.PasswordHash,
	)

	return scanUser(row)
}

func (queries *Queries) GetUserByID(
	ctx context.Context,
	id uuid.UUID,
) (User, error) {
	row := queries.db.QueryRow(
		ctx,
		`
			SELECT
				id,
				name,
				email,
				password_hash,
				created_at,
				updated_at
			FROM users
			WHERE id = $1
			LIMIT 1
		`,
		id,
	)

	return scanUser(row)
}

func (queries *Queries) GetUserByEmail(
	ctx context.Context,
	email string,
) (User, error) {
	row := queries.db.QueryRow(
		ctx,
		`
			SELECT
				id,
				name,
				email,
				password_hash,
				created_at,
				updated_at
			FROM users
			WHERE LOWER(email) = LOWER($1)
			LIMIT 1
		`,
		email,
	)

	return scanUser(row)
}

func (queries *Queries) CreateSession(
	ctx context.Context,
	params CreateSessionParams,
) (Session, error) {
	row := queries.db.QueryRow(
		ctx,
		`
			INSERT INTO sessions (
				id,
				user_id,
				token_hash,
				csrf_hash,
				user_agent,
				ip_address,
				expires_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING
				id,
				user_id,
				token_hash,
				csrf_hash,
				user_agent,
				ip_address,
				expires_at,
				last_used_at,
				created_at
		`,
		params.ID,
		params.UserID,
		params.TokenHash,
		params.CSRFHash,
		params.UserAgent,
		params.IPAddress,
		params.ExpiresAt,
	)

	return scanSession(row)
}

func (queries *Queries) GetSessionByTokenHash(
	ctx context.Context,
	tokenHash string,
) (Session, error) {
	row := queries.db.QueryRow(
		ctx,
		`
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
			WHERE token_hash = $1
			  AND expires_at > NOW()
			LIMIT 1
		`,
		tokenHash,
	)

	return scanSession(row)
}

func (queries *Queries) TouchSession(
	ctx context.Context,
	id uuid.UUID,
) error {
	_, err := queries.db.Exec(
		ctx,
		`
			UPDATE sessions
			SET last_used_at = NOW()
			WHERE id = $1
		`,
		id,
	)

	return err
}

func (queries *Queries) DeleteSessionByTokenHash(
	ctx context.Context,
	tokenHash string,
) error {
	_, err := queries.db.Exec(
		ctx,
		`
			DELETE FROM sessions
			WHERE token_hash = $1
		`,
		tokenHash,
	)

	return err
}

func scanUser(row interface {
	Scan(...any) error
}) (User, error) {
	var user User

	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	return user, err
}

func scanSession(row interface {
	Scan(...any) error
}) (Session, error) {
	var session Session

	err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&session.CSRFHash,
		&session.UserAgent,
		&session.IPAddress,
		&session.ExpiresAt,
		&session.LastUsedAt,
		&session.CreatedAt,
	)

	return session, err
}
