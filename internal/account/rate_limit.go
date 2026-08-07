package account

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

func (handler *Handler) allowRequest(
	ctx context.Context,
	key string,
	limit int64,
	window time.Duration,
) (bool, error) {
	hash := sha256.Sum256([]byte(key))

	redisKey := "rate-limit:" +
		hex.EncodeToString(hash[:])

	count, err := handler.redis.Incr(
		ctx,
		redisKey,
	).Result()

	if err != nil {
		return false, err
	}

	if count == 1 {
		if err := handler.redis.Expire(
			ctx,
			redisKey,
			window,
		).Err(); err != nil {
			return false, err
		}
	}

	return count <= limit, nil
}
