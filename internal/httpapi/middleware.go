package httpapi

import (
	"context"
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
				next.ServeHTTP(writer, request)
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

			next.ServeHTTP(writer, request)
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

			next.ServeHTTP(writer, request)
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

			if request.Method ==
				http.MethodOptions {
				writer.WriteHeader(
					http.StatusNoContent,
				)
				return
			}

			next.ServeHTTP(writer, request)
		},
	)
}

func (server *Server) allowRequest(
	ctx context.Context,
	key string,
	limit int64,
	window time.Duration,
) (bool, error) {
	rateKey := "rate:" + key

	count, err := server.redisClient.Incr(
		ctx,
		rateKey,
	).Result()
	if err != nil {
		return false, err
	}

	if count == 1 {
		if err := server.redisClient.Expire(
			ctx,
			rateKey,
			window,
		).Err(); err != nil {
			return false, err
		}
	}

	return count <= limit, nil
}
