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
	keyHash := hex.EncodeToString(hash[:])

	windowSeconds := window.Seconds()
	if windowSeconds < 1 {
		windowSeconds = 1
	}

	var count int64

	err := handler.pool.QueryRow(
		ctx,
		`
			INSERT INTO rate_limits (
				key_hash,
				request_count,
				window_started_at,
				expires_at
			)
			VALUES (
				$1,
				1,
				NOW(),
				NOW() + make_interval(
					secs => $2::double precision
				)
			)
			ON CONFLICT (key_hash)
			DO UPDATE SET
				request_count = CASE
					WHEN rate_limits.expires_at <= NOW()
					THEN 1
					ELSE rate_limits.request_count + 1
				END,
				window_started_at = CASE
					WHEN rate_limits.expires_at <= NOW()
					THEN NOW()
					ELSE rate_limits.window_started_at
				END,
				expires_at = CASE
					WHEN rate_limits.expires_at <= NOW()
					THEN NOW() + make_interval(
						secs => $2::double precision
					)
					ELSE rate_limits.expires_at
				END
			RETURNING request_count
		`,
		keyHash,
		windowSeconds,
	).Scan(&count)

	if err != nil {
		return false, err
	}

	return count <= limit, nil
}
