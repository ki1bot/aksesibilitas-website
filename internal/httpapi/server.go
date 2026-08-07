package httpapi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/ki1bot/aksesibilitas-website/internal/auth"
	"github.com/ki1bot/aksesibilitas-website/internal/config"
	db "github.com/ki1bot/aksesibilitas-website/internal/database/db"
)

type principalContextKey struct{}

type Server struct {
	cfg         config.Config
	pool        *pgxpool.Pool
	queries     *db.Queries
	redisClient *redis.Client
	queueClient *asynq.Client
	sessions    *auth.Manager
}

func NewRouter(
	cfg config.Config,
	pool *pgxpool.Pool,
	redisClient *redis.Client,
	queueClient *asynq.Client,
) http.Handler {
	queries := db.New(pool)

	server := &Server{
		cfg:         cfg,
		pool:        pool,
		queries:     queries,
		redisClient: redisClient,
		queueClient: queueClient,
		sessions: auth.NewManager(
			queries,
			redisClient,
			cfg.SessionCookieName,
			cfg.SessionTTL,
			cfg.AppEnv == "production",
		),
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(90 * time.Second))
	router.Use(server.securityHeaders)
	router.Use(server.cors)

	router.Route("/api/v1", server.apiRoutes)

	return router
}
