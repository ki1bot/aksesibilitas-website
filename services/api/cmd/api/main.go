package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"

	"github.com/ki1bot/aksescheck-id/internal/config"
	"github.com/ki1bot/aksescheck-id/internal/database"
	"github.com/ki1bot/aksescheck-id/internal/httpapi"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	rootContext := context.Background()

	pool, err := database.Open(rootContext, cfg.DatabaseURL)
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

	queueClient := asynq.NewClient(
		asynq.RedisClientOpt{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		},
	)
	defer queueClient.Close()

	handler := httpapi.NewRouter(
		cfg,
		pool,
		redisClient,
		queueClient,
	)

	server := &http.Server{
		Addr:              cfg.APIAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	signalContext, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	errorChannel := make(chan error, 1)

	go func() {
		log.Printf("AksesCheck API berjalan pada %s", cfg.APIAddr)
		errorChannel <- server.ListenAndServe()
	}()

	select {
	case <-signalContext.Done():
	case serverError := <-errorChannel:
		if !errors.Is(serverError, http.ErrServerClosed) {
			log.Fatal(serverError)
		}
		return
	}

	shutdownContext, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownContext); err != nil {
		log.Printf("gagal menghentikan server dengan bersih: %v", err)
	}
}
