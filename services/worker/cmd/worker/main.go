package main

import (
	"context"
	"log"
	"time"

	"github.com/hibiken/asynq"

	"github.com/ki1bot/aksescheck-id/internal/config"
	"github.com/ki1bot/aksescheck-id/internal/database"
	db "github.com/ki1bot/aksescheck-id/internal/database/db"
	taskqueue "github.com/ki1bot/aksescheck-id/internal/queue"
	"github.com/ki1bot/aksescheck-id/internal/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	pool, err := database.Open(
		context.Background(),
		cfg.DatabaseURL,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	server := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		},
		asynq.Config{
			Concurrency: cfg.WorkerConcurrency,
			Queues: map[string]int{
				cfg.ScanQueue: 1,
			},
			ShutdownTimeout: 30 * time.Second,
		},
	)

	mux := asynq.NewServeMux()

	mux.Handle(
		taskqueue.TypeAccessibilityScan,
		worker.NewScanHandler(db.New(pool)),
	)

	log.Printf(
		"AksesCheck worker berjalan pada antrean %s",
		cfg.ScanQueue,
	)

	if err := server.Run(mux); err != nil {
		log.Fatal(err)
	}
}
