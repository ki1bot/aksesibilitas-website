package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"

	"github.com/ki1bot/aksesibilitas-website/internal/config"
	"github.com/ki1bot/aksesibilitas-website/internal/database"
	taskqueue "github.com/ki1bot/aksesibilitas-website/internal/queue"
	"github.com/ki1bot/aksesibilitas-website/internal/worker"
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

	redisOptions := asynq.RedisClientOpt{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}

	server := asynq.NewServer(
		redisOptions,
		asynq.Config{
			Concurrency: cfg.WorkerConcurrency,
			Queues: map[string]int{
				cfg.ScanQueue: 1,
			},
			StrictPriority: true,
		},
	)

	mux := asynq.NewServeMux()

	scanHandler := worker.NewScanHandler(
		pool,
		cfg.ChromePath,
	)

	mux.HandleFunc(
		taskqueue.AccessibilityScanTask,
		scanHandler.ProcessTask,
	)

	log.Printf(
		"AksesCheck ID worker berjalan dengan concurrency %d",
		cfg.WorkerConcurrency,
	)

	if err := server.Run(mux); err != nil {
		log.Fatal(err)
	}
}
