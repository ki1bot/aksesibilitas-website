package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/ki1bot/aksesibilitas-website/internal/database/db"
)

var ErrUnauthenticated = errors.New(
	"sesi tidak valid atau sudah berakhir",
)

var ErrInvalidCSRF = errors.New(
	"token CSRF tidak valid",
)

type Manager struct {
	queries    *db.Queries
	cookieName string
	ttl        time.Duration
	secure     bool
}

type Principal struct {
	User      db.User
	SessionID uuid.UUID
	TokenHash string
	CSRFHash  string
	ExpiresAt time.Time
}

type Tokens struct {
	SessionToken string
	CSRFToken    string
	ExpiresAt    time.Time
}

func NewManager(
	queries *db.Queries,
	cookieName string,
	ttl time.Duration,
	secure bool,
) *Manager {
	return &Manager{
		queries:    queries,
		cookieName: cookieName,
		ttl:        ttl,
		secure:     secure,
	}
}

func (manager *Manager) Create(
	ctx context.Context,
	userID uuid.UUID,
	userAgent string,
	ipAddress string,
) (Tokens, error) {
	sessionToken, err := randomToken(32)
	if err != nil {
		return Tokens{}, err
	}

	csrfToken, err := randomToken(32)
	if err != nil {
		return Tokens{}, err
	}

	expiresAt := time.Now().Add(manager.ttl)

	_, err = manager.queries.CreateSession(
		ctx,
		db.CreateSessionParams{
			ID:        uuid.New(),
			UserID:    userID,
			TokenHash: hashToken(sessionToken),
			CSRFHash:  hashToken(csrfToken),
			UserAgent: userAgent,
			IPAddress: ipAddress,
			ExpiresAt: expiresAt,
		},
	)
	if err != nil {
		return Tokens{}, err
	}

	return Tokens{
		SessionToken: sessionToken,
		CSRFToken:    csrfToken,
		ExpiresAt:    expiresAt,
	}, nil
}

func (manager *Manager) Authenticate(
	ctx context.Context,
	request *http.Request,
) (Principal, error) {
	cookie, err := request.Cookie(manager.cookieName)
	if err != nil || cookie.Value == "" {
		return Principal{}, ErrUnauthenticated
	}

	tokenHash := hashToken(cookie.Value)

	session, err := manager.queries.GetSessionByTokenHash(
		ctx,
		tokenHash,
	)
	if err != nil {
		return Principal{}, ErrUnauthenticated
	}

	if time.Now().After(session.ExpiresAt) {
		_ = manager.DestroyByHash(ctx, tokenHash)
		return Principal{}, ErrUnauthenticated
	}

	user, err := manager.queries.GetUserByID(
		ctx,
		session.UserID,
	)
	if err != nil {
		return Principal{}, ErrUnauthenticated
	}

	_ = manager.queries.TouchSession(ctx, session.ID)

	return Principal{
		User:      user,
		SessionID: session.ID,
		TokenHash: tokenHash,
		CSRFHash:  session.CSRFHash,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

func (manager *Manager) ValidateCSRF(
	principal Principal,
	token string,
) error {
	if token == "" {
		return ErrInvalidCSRF
	}

	expected, err := hex.DecodeString(principal.CSRFHash)
	if err != nil {
		return ErrInvalidCSRF
	}

	actualHash := sha256.Sum256([]byte(token))

	if subtle.ConstantTimeCompare(
		expected,
		actualHash[:],
	) != 1 {
		return ErrInvalidCSRF
	}

	return nil
}

func (manager *Manager) Destroy(
	ctx context.Context,
	request *http.Request,
) error {
	cookie, err := request.Cookie(manager.cookieName)
	if err != nil || cookie.Value == "" {
		return nil
	}

	return manager.DestroyByHash(
		ctx,
		hashToken(cookie.Value),
	)
}

func (manager *Manager) DestroyByHash(
	ctx context.Context,
	tokenHash string,
) error {
	err := manager.queries.DeleteSessionByTokenHash(
		ctx,
		tokenHash,
	)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	return nil
}

func (manager *Manager) SetCookies(
	writer http.ResponseWriter,
	token string,
	csrfToken string,
	expiresAt time.Time,
) {
	http.SetCookie(
		writer,
		&http.Cookie{
			Name:     manager.cookieName,
			Value:    token,
			Path:     "/",
			Expires:  expiresAt,
			MaxAge:   int(time.Until(expiresAt).Seconds()),
			HttpOnly: true,
			Secure:   manager.secure,
			SameSite: http.SameSiteLaxMode,
		},
	)

	http.SetCookie(
		writer,
		&http.Cookie{
			Name:     manager.cookieName + "_csrf",
			Value:    csrfToken,
			Path:     "/",
			Expires:  expiresAt,
			MaxAge:   int(time.Until(expiresAt).Seconds()),
			HttpOnly: false,
			Secure:   manager.secure,
			SameSite: http.SameSiteLaxMode,
		},
	)
}

func (manager *Manager) ClearCookie(
	writer http.ResponseWriter,
) {
	http.SetCookie(
		writer,
		&http.Cookie{
			Name:     manager.cookieName,
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   manager.secure,
			SameSite: http.SameSiteLaxMode,
		},
	)

	http.SetCookie(
		writer,
		&http.Cookie{
			Name:     manager.cookieName + "_csrf",
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
			HttpOnly: false,
			Secure:   manager.secure,
			SameSite: http.SameSiteLaxMode,
		},
	)
}

func randomToken(size int) (string, error) {
	value := make([]byte, size)

	if _, err := rand.Read(value); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(value), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))

	return hex.EncodeToString(hash[:])
}
