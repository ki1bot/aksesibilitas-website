package account

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (handler *Handler) authenticate(
	ctx context.Context,
	request *http.Request,
) (authenticatedUser, error) {
	cookie, err := request.Cookie(
		handler.cfg.SessionCookieName,
	)
	if err != nil ||
		strings.TrimSpace(cookie.Value) == "" {
		return authenticatedUser{}, pgx.ErrNoRows
	}

	var user authenticatedUser

	err = handler.pool.QueryRow(
		ctx,
		`
			SELECT users.id, users.password_hash
			FROM sessions
			JOIN users ON users.id = sessions.user_id
			WHERE sessions.token_hash = $1
			  AND sessions.expires_at > NOW()
			LIMIT 1
		`,
		hashToken(cookie.Value),
	).Scan(
		&user.ID,
		&user.PasswordHash,
	)

	return user, err
}

func (handler *Handler) validCSRF(
	request *http.Request,
) bool {
	headerToken := strings.TrimSpace(
		request.Header.Get("X-CSRF-Token"),
	)
	if headerToken == "" {
		return false
	}

	csrfCookie, err := request.Cookie(
		handler.cfg.SessionCookieName + "_csrf",
	)
	if err != nil || csrfCookie.Value == "" {
		return false
	}

	if subtle.ConstantTimeCompare(
		[]byte(headerToken),
		[]byte(csrfCookie.Value),
	) != 1 {
		return false
	}

	sessionCookie, err := request.Cookie(
		handler.cfg.SessionCookieName,
	)
	if err != nil || sessionCookie.Value == "" {
		return false
	}

	var csrfHash string

	err = handler.pool.QueryRow(
		request.Context(),
		`
			SELECT csrf_hash
			FROM sessions
			WHERE token_hash = $1
			  AND expires_at > NOW()
			LIMIT 1
		`,
		hashToken(sessionCookie.Value),
	).Scan(&csrfHash)

	if err != nil {
		return false
	}

	expected, err := hex.DecodeString(csrfHash)
	if err != nil {
		return false
	}

	actual := sha256.Sum256(
		[]byte(headerToken),
	)

	return subtle.ConstantTimeCompare(
		expected,
		actual[:],
	) == 1
}

func (handler *Handler) replacePasswordAndInvalidateSessions(
	ctx context.Context,
	userID uuid.UUID,
	passwordHash string,
) ([]string, error) {
	transaction, err := handler.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer transaction.Rollback(ctx)

	rows, err := transaction.Query(
		ctx,
		`
			SELECT token_hash
			FROM sessions
			WHERE user_id = $1
		`,
		userID,
	)
	if err != nil {
		return nil, err
	}

	sessionHashes := make([]string, 0)

	for rows.Next() {
		var tokenHash string

		if err := rows.Scan(
			&tokenHash,
		); err != nil {
			rows.Close()
			return nil, err
		}

		sessionHashes = append(
			sessionHashes,
			tokenHash,
		)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}

	rows.Close()

	commandTag, err := transaction.Exec(
		ctx,
		`
			UPDATE users
			SET password_hash = $1,
			    updated_at = NOW()
			WHERE id = $2
		`,
		passwordHash,
		userID,
	)
	if err != nil {
		return nil, err
	}

	if commandTag.RowsAffected() != 1 {
		return nil, pgx.ErrNoRows
	}

	if _, err := transaction.Exec(
		ctx,
		`
			DELETE FROM sessions
			WHERE user_id = $1
		`,
		userID,
	); err != nil {
		return nil, err
	}

	if err := transaction.Commit(ctx); err != nil {
		return nil, err
	}

	return sessionHashes, nil
}

func (handler *Handler) removeSessionCache(
	ctx context.Context,
	sessionHashes []string,
) {
	if len(sessionHashes) == 0 {
		return
	}

	keys := make(
		[]string,
		0,
		len(sessionHashes),
	)

	for _, tokenHash := range sessionHashes {
		keys = append(
			keys,
			"session:"+tokenHash,
		)
	}

	if err := handler.redis.Del(
		ctx,
		keys...,
	).Err(); err != nil {
		log.Printf(
			"gagal menghapus cache sesi: %v",
			err,
		)
	}
}

func (handler *Handler) clearCookies(
	writer http.ResponseWriter,
) {
	secure := handler.cfg.AppEnv == "production"

	for _, cookie := range []*http.Cookie{
		{
			Name:  handler.cfg.SessionCookieName,
			Value: "",
			Path:  "/",
			Expires: time.Unix(
				0,
				0,
			),
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   secure,
			SameSite: http.SameSiteLaxMode,
		},
		{
			Name: handler.cfg.SessionCookieName +
				"_csrf",
			Value: "",
			Path:  "/",
			Expires: time.Unix(
				0,
				0,
			),
			MaxAge:   -1,
			HttpOnly: false,
			Secure:   secure,
			SameSite: http.SameSiteLaxMode,
		},
	} {
		http.SetCookie(writer, cookie)
	}
}
