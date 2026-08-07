package httpapi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ki1bot/aksesibilitas-website/internal/account"
	"github.com/ki1bot/aksesibilitas-website/internal/auth"
	"github.com/ki1bot/aksesibilitas-website/internal/config"
	db "github.com/ki1bot/aksesibilitas-website/internal/database/db"
)

type principalContextKey struct{}

type Server struct {
	cfg        config.Config
	pool       *pgxpool.Pool
	queries    *db.Queries
	sessions   *auth.Manager
	httpClient *http.Client
}

func NewRouter(
	cfg config.Config,
	pool *pgxpool.Pool,
) http.Handler {
	queries := db.New(pool)

	server := &Server{
		cfg:     cfg,
		pool:    pool,
		queries: queries,
		sessions: auth.NewManager(
			queries,
			cfg.SessionCookieName,
			cfg.SessionTTL,
			cfg.AppEnv == "production",
		),
		httpClient: &http.Client{
			Timeout: cfg.ScanTimeout + 15*time.Second,
		},
	}

	passwordHandler := account.NewHandler(
		cfg,
		pool,
	)

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(
		middleware.Timeout(
			cfg.ScanTimeout + 30*time.Second,
		),
	)
	router.Use(server.securityHeaders)
	router.Use(server.cors)

	for _, path := range []string{
		"/api/v1/auth/forgot-password",
		"/api/v1/auth/reset-password",
		"/api/v1/auth/change-password",
	} {
		router.Options(
			path,
			passwordHandler.Options,
		)
	}

	router.Post(
		"/api/v1/auth/forgot-password",
		passwordHandler.ForgotPassword,
	)

	router.Post(
		"/api/v1/auth/reset-password",
		passwordHandler.ResetPassword,
	)

	router.Post(
		"/api/v1/auth/change-password",
		passwordHandler.ChangePassword,
	)

	router.Route(
		"/api/v1",
		server.apiRoutes,
	)

	return router
}
