package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

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

	router := httpapi.NewRouter(
		cfg,
		pool,
	)

	server := &http.Server{
		Addr:              cfg.APIAddr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      cfg.ScanTimeout + 30*time.Second,
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

	shutdownContext, cancelShutdown := context.WithTimeout(
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
