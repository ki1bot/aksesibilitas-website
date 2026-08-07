package account

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func (handler *Handler) storeResetToken(
	ctx context.Context,
	userID uuid.UUID,
	tokenHash string,
) error {
	userKey := handler.resetUserKey(userID)

	previousHash, err := handler.redis.Get(
		ctx,
		userKey,
	).Result()

	if err != nil &&
		!errors.Is(err, redis.Nil) {
		return err
	}

	pipeline := handler.redis.TxPipeline()

	if previousHash != "" {
		pipeline.Del(
			ctx,
			handler.resetTokenKey(previousHash),
		)
	}

	pipeline.Set(
		ctx,
		handler.resetTokenKey(tokenHash),
		userID.String(),
		handler.cfg.PasswordResetTTL,
	)

	pipeline.Set(
		ctx,
		userKey,
		tokenHash,
		handler.cfg.PasswordResetTTL,
	)

	_, err = pipeline.Exec(ctx)
	return err
}

func (handler *Handler) resetTokenKey(
	tokenHash string,
) string {
	return "password-reset:" + tokenHash
}

func (handler *Handler) resetUserKey(
	userID uuid.UUID,
) string {
	return "password-reset-user:" +
		userID.String()
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
