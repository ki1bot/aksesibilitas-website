package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"
)

func (server *Server) authenticate(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(
			writer http.ResponseWriter,
			request *http.Request,
		) {
			principal, err :=
				server.sessions.Authenticate(
					request.Context(),
					request,
				)
			if err != nil {
				writeError(
					writer,
					http.StatusUnauthorized,
					"unauthenticated",
					"Silakan login terlebih dahulu",
				)
				return
			}

			contextWithPrincipal :=
				context.WithValue(
					request.Context(),
					principalContextKey{},
					principal,
				)

			next.ServeHTTP(
				writer,
				request.WithContext(
					contextWithPrincipal,
				),
			)
		},
	)
}

func (server *Server) csrf(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(
			writer http.ResponseWriter,
			request *http.Request,
		) {
			if request.Method == http.MethodGet ||
				request.Method == http.MethodHead ||
				request.Method == http.MethodOptions {
				next.ServeHTTP(
					writer,
					request,
				)
				return
			}

			principal := principalFromContext(
				request.Context(),
			)

			err := server.sessions.ValidateCSRF(
				principal,
				request.Header.Get("X-CSRF-Token"),
			)
			if err != nil {
				writeError(
					writer,
					http.StatusForbidden,
					"invalid_csrf",
					"Token CSRF tidak valid",
				)
				return
			}

			next.ServeHTTP(
				writer,
				request,
			)
		},
	)
}

func (server *Server) securityHeaders(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(
			writer http.ResponseWriter,
			request *http.Request,
		) {
			writer.Header().Set(
				"X-Content-Type-Options",
				"nosniff",
			)

			writer.Header().Set(
				"X-Frame-Options",
				"DENY",
			)

			writer.Header().Set(
				"Referrer-Policy",
				"strict-origin-when-cross-origin",
			)

			writer.Header().Set(
				"Permissions-Policy",
				"camera=(), microphone=(), geolocation=()",
			)

			next.ServeHTTP(
				writer,
				request,
			)
		},
	)
}

func (server *Server) cors(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(
			writer http.ResponseWriter,
			request *http.Request,
		) {
			origin := request.Header.Get("Origin")

			if origin != "" &&
				origin == server.cfg.WebOrigin {
				writer.Header().Set(
					"Access-Control-Allow-Origin",
					origin,
				)

				writer.Header().Set(
					"Access-Control-Allow-Credentials",
					"true",
				)

				writer.Header().Add(
					"Vary",
					"Origin",
				)
			}

			writer.Header().Set(
				"Access-Control-Allow-Headers",
				"Accept, Content-Type, X-CSRF-Token",
			)

			writer.Header().Set(
				"Access-Control-Allow-Methods",
				"GET, POST, PATCH, DELETE, OPTIONS",
			)

			if request.Method == http.MethodOptions {
				writer.WriteHeader(
					http.StatusNoContent,
				)
				return
			}

			next.ServeHTTP(
				writer,
				request,
			)
		},
	)
}

func (server *Server) allowRequest(
	ctx context.Context,
	key string,
	limit int64,
	window time.Duration,
) (bool, error) {
	hash := sha256.Sum256(
		[]byte(key),
	)

	keyHash := hex.EncodeToString(
		hash[:],
	)

	windowSeconds := int64(
		window / time.Second,
	)

	if windowSeconds < 1 {
		windowSeconds = 1
	}

	var count int64

	err := server.pool.QueryRow(
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
				NOW() + (
					$2::bigint * INTERVAL '1 second'
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
					THEN NOW() + (
						$2::bigint * INTERVAL '1 second'
					)
					ELSE rate_limits.expires_at
				END
			RETURNING request_count
		`,
		keyHash,
		windowSeconds,
	).Scan(
		&count,
	)

	if err != nil {
		return false, err
	}

	return count <= limit, nil
}
