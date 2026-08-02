package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"

	"github.com/ki1bot/aksesibilitas-website/internal/database/db"
)

var ErrUnauthenticated = errors.New("sesi tidak valid atau sudah berakhir")
var ErrInvalidCSRF = errors.New("token CSRF tidak valid")

type Manager struct {
	queries    *db.Queries
	redis      *redis.Client
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

type cachedSession struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	CSRFHash  string    `json:"csrf_hash"`
	ExpiresAt time.Time `json:"expires_at"`
}

func NewManager(
	queries *db.Queries,
	redisClient *redis.Client,
	cookieName string,
	ttl time.Duration,
	secure bool,
) *Manager {
	return &Manager{
		queries:    queries,
		redis:      redisClient,
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

	session, err := manager.queries.CreateSession(
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

	if err := manager.cacheSession(ctx, session); err != nil {
		_ = manager.queries.DeleteSessionByTokenHash(
			ctx,
			session.TokenHash,
		)
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

	session, err := manager.loadSession(ctx, tokenHash)
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

	return manager.redis.Del(
		ctx,
		manager.cacheKey(tokenHash),
	).Err()
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

func (manager *Manager) loadSession(
	ctx context.Context,
	tokenHash string,
) (db.Session, error) {
	value, err := manager.redis.Get(
		ctx,
		manager.cacheKey(tokenHash),
	).Bytes()

	if err == nil {
		var cached cachedSession

		if json.Unmarshal(value, &cached) == nil {
			return db.Session{
				ID:        cached.ID,
				UserID:    cached.UserID,
				TokenHash: tokenHash,
				CSRFHash:  cached.CSRFHash,
				ExpiresAt: cached.ExpiresAt,
			}, nil
		}
	}

	session, err := manager.queries.GetSessionByTokenHash(
		ctx,
		tokenHash,
	)
	if err != nil {
		return db.Session{}, err
	}

	if err := manager.cacheSession(ctx, session); err != nil {
		return db.Session{}, err
	}

	return session, nil
}

func (manager *Manager) cacheSession(
	ctx context.Context,
	session db.Session,
) error {
	payload, err := json.Marshal(
		cachedSession{
			ID:        session.ID,
			UserID:    session.UserID,
			CSRFHash:  session.CSRFHash,
			ExpiresAt: session.ExpiresAt,
		},
	)
	if err != nil {
		return err
	}

	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return ErrUnauthenticated
	}

	return manager.redis.Set(
		ctx,
		manager.cacheKey(session.TokenHash),
		payload,
		ttl,
	).Err()
}

func (manager *Manager) cacheKey(
	tokenHash string,
) string {
	return "session:" + tokenHash
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
