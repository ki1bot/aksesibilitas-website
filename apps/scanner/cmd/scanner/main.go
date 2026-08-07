package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ki1bot/aksesibilitas-website/internal/config"
	"github.com/ki1bot/aksesibilitas-website/internal/database"
	"github.com/ki1bot/aksesibilitas-website/internal/worker"
)

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

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

	scanHandler := worker.NewScanHandler(
		pool,
		cfg.ChromePath,
	)

	router := chi.NewRouter()

	router.Get(
		"/internal/health",
		func(
			writer http.ResponseWriter,
			request *http.Request,
		) {
			writeJSON(
				writer,
				http.StatusOK,
				map[string]string{
					"status": "ok",
				},
			)
		},
	)

	router.Post(
		"/internal/scans/{scanId}",
		func(
			writer http.ResponseWriter,
			request *http.Request,
		) {
			if !authorized(
				request,
				cfg.ScannerToken,
			) {
				writeJSON(
					writer,
					http.StatusUnauthorized,
					errorResponse{
						Code:    "unauthorized",
						Message: "Akses scanner ditolak",
					},
				)
				return
			}

			scanID, err := uuid.Parse(
				chi.URLParam(
					request,
					"scanId",
				),
			)
			if err != nil {
				writeJSON(
					writer,
					http.StatusBadRequest,
					errorResponse{
						Code:    "invalid_scan_id",
						Message: "Format scan ID tidak valid",
					},
				)
				return
			}

			scanContext, cancel := context.WithTimeout(
				request.Context(),
				cfg.ScanTimeout,
			)
			defer cancel()

			err = scanHandler.Process(
				scanContext,
				scanID,
			)
			if err != nil {
				log.Printf(
					"scanner gagal memproses %s: %v",
					scanID,
					err,
				)

				writeJSON(
					writer,
					http.StatusInternalServerError,
					errorResponse{
						Code:    "scan_failed",
						Message: "Pemindaian gagal diselesaikan",
					},
				)
				return
			}

			writer.WriteHeader(http.StatusNoContent)
		},
	)

	server := &http.Server{
		Addr:              cfg.ScannerAddr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      cfg.ScanTimeout + 30*time.Second,
		IdleTimeout:       2 * time.Minute,
	}

	go func() {
		log.Printf(
			"AksesCheck ID scanner berjalan pada %s",
			cfg.ScannerAddr,
		)

		err := server.ListenAndServe()

		if err != nil &&
			!errors.Is(
				err,
				http.ErrServerClosed,
			) {
			log.Fatalf(
				"scanner berhenti: %v",
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
			"scanner gagal berhenti dengan bersih: %v",
			err,
		)
	}
}

func authorized(
	request *http.Request,
	expectedToken string,
) bool {
	header := strings.TrimSpace(
		request.Header.Get("Authorization"),
	)

	const prefix = "Bearer "

	if !strings.HasPrefix(header, prefix) {
		return false
	}

	actualToken := strings.TrimSpace(
		strings.TrimPrefix(
			header,
			prefix,
		),
	)

	if actualToken == "" || expectedToken == "" {
		return false
	}

	return subtle.ConstantTimeCompare(
		[]byte(actualToken),
		[]byte(expectedToken),
	) == 1
}

func writeJSON(
	writer http.ResponseWriter,
	status int,
	payload any,
) {
	writer.Header().Set(
		"Content-Type",
		"application/json; charset=utf-8",
	)

	writer.WriteHeader(status)

	_ = json.NewEncoder(writer).Encode(payload)
}
