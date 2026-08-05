package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"

	"github.com/ki1bot/aksesibilitas-website/internal/account"
	"github.com/ki1bot/aksesibilitas-website/internal/config"
	"github.com/ki1bot/aksesibilitas-website/internal/database"
	"github.com/ki1bot/aksesibilitas-website/internal/httpapi"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	pool, err := database.Open(
		ctx,
		cfg.DatabaseURL,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	redisClient := redis.NewClient(
		&redis.Options{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		},
	)
	defer redisClient.Close()

	pingContext, cancelPing := context.WithTimeout(
		ctx,
		5*time.Second,
	)

	if err := redisClient.Ping(
		pingContext,
	).Err(); err != nil {
		cancelPing()

		log.Fatalf(
			"Redis tidak dapat dihubungi: %v",
			err,
		)
	}

	cancelPing()

	queueClient := asynq.NewClient(
		asynq.RedisClientOpt{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		},
	)
	defer queueClient.Close()

	baseRouter := httpapi.NewRouter(
		cfg,
		pool,
		redisClient,
		queueClient,
	)

	passwordHandler := account.NewHandler(
		cfg,
		pool,
		redisClient,
	)

	router := chi.NewRouter()

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

	router.NotFound(baseRouter.ServeHTTP)
	router.MethodNotAllowed(baseRouter.ServeHTTP)

	server := &http.Server{
		Addr:              cfg.APIAddr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      90 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}

	go func() {
		log.Printf(
			"AksesCheck ID API berjalan pada %s",
			cfg.APIAddr,
		)

		err := server.ListenAndServe()

		if err != nil &&
			!errors.Is(
				err,
				http.ErrServerClosed,
			) {
			log.Fatalf(
				"API berhenti: %v",
				err,
			)
		}
	}()

	<-ctx.Done()

	shutdownContext, cancelShutdown :=
		context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
	defer cancelShutdown()

	if err := server.Shutdown(
		shutdownContext,
	); err != nil {
		log.Printf(
			"API gagal berhenti dengan bersih: %v",
			err,
		)
	}
}
