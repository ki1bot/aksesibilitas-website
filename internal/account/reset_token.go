package account

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

func (handler *Handler) storeResetToken(
	ctx context.Context,
	userID uuid.UUID,
	tokenHash string,
) error {
	_, err := handler.pool.Exec(
		ctx,
		`
			INSERT INTO password_reset_tokens (
				token_hash,
				user_id,
				expires_at
			)
			VALUES (
				$1,
				$2,
				$3
			)
			ON CONFLICT (user_id)
			DO UPDATE SET
				token_hash = EXCLUDED.token_hash,
				expires_at = EXCLUDED.expires_at,
				created_at = NOW()
		`,
		tokenHash,
		userID,
		time.Now().Add(handler.cfg.PasswordResetTTL),
	)

	return err
}

func (handler *Handler) consumeResetToken(
	ctx context.Context,
	tokenHash string,
) (uuid.UUID, error) {
	var userID uuid.UUID

	err := handler.pool.QueryRow(
		ctx,
		`
			DELETE FROM password_reset_tokens
			WHERE token_hash = $1
			  AND expires_at > NOW()
			RETURNING user_id
		`,
		tokenHash,
	).Scan(&userID)

	return userID, err
}

func randomToken(
	size int,
) (string, error) {
	value := make([]byte, size)

	if _, err := rand.Read(value); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(
		value,
	), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))

	return hex.EncodeToString(hash[:])
}
