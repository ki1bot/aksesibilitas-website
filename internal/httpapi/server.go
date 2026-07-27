package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/ki1bot/aksescheck-id/internal/config"
	db "github.com/ki1bot/aksescheck-id/internal/database/db"
	taskqueue "github.com/ki1bot/aksescheck-id/internal/queue"
	"github.com/ki1bot/aksescheck-id/internal/security"
)

type Server struct {
	cfg         config.Config
	pool        *pgxpool.Pool
	queries     db.Querier
	redisClient *redis.Client
	queueClient *asynq.Client
}

type healthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Redis    string `json:"redis"`
}

type createScanRequest struct {
	URL string `json:"url"`
}

type scanResponse struct {
	ID           string    `json:"id"`
	URL          string    `json:"url"`
	Status       string    `json:"status"`
	ErrorMessage string    `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewRouter(
	cfg config.Config,
	pool *pgxpool.Pool,
	redisClient *redis.Client,
	queueClient *asynq.Client,
) http.Handler {
	server := &Server{
		cfg:         cfg,
		pool:        pool,
		queries:     db.New(pool),
		redisClient: redisClient,
		queueClient: queueClient,
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(30 * time.Second))
	router.Use(server.cors)

	router.Route("/api/v1", func(router chi.Router) {
		router.Get("/health", server.health)
		router.Post("/scans", server.createScan)
		router.Get("/scans/{scanId}", server.getScan)
	})

	return router
}

func (server *Server) health(
	writer http.ResponseWriter,
	request *http.Request,
) {
	response := healthResponse{
		Status:   "ok",
		Database: "ok",
		Redis:    "ok",
	}

	statusCode := http.StatusOK

	if err := server.pool.Ping(request.Context()); err != nil {
		response.Status = "degraded"
		response.Database = "unavailable"
		statusCode = http.StatusServiceUnavailable
	}

	if err := server.redisClient.Ping(request.Context()).Err(); err != nil {
		response.Status = "degraded"
		response.Redis = "unavailable"
		statusCode = http.StatusServiceUnavailable
	}

	writeJSON(writer, statusCode, response)
}

func (server *Server) createScan(
	writer http.ResponseWriter,
	request *http.Request,
) {
	var input createScanRequest

	if err := readJSON(writer, request, &input); err != nil {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_request",
			"Body JSON tidak valid",
		)
		return
	}

	normalizedURL, err := security.ValidatePublicHTTPURL(
		request.Context(),
		input.URL,
	)
	if err != nil {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_url",
			err.Error(),
		)
		return
	}

	scan, err := server.queries.CreateScan(
		request.Context(),
		db.CreateScanParams{
			ID:  uuid.New(),
			URL: normalizedURL,
		},
	)
	if err != nil {
		log.Printf("gagal membuat scan: %v", err)

		writeError(
			writer,
			http.StatusInternalServerError,
			"database_error",
			"Gagal membuat data scan",
		)
		return
	}

	task, err := taskqueue.NewAccessibilityScanTask(
		scan.ID.String(),
		scan.URL,
		server.cfg.ScanQueue,
		server.cfg.ScanTimeout,
	)
	if err != nil {
		server.markScanFailed(
			request,
			scan.ID,
			"Gagal membuat payload antrean",
		)

		writeError(
			writer,
			http.StatusInternalServerError,
			"queue_payload_error",
			"Gagal membuat pekerjaan scan",
		)
		return
	}

	if _, err := server.queueClient.Enqueue(task); err != nil {
		log.Printf("gagal memasukkan task ke Redis: %v", err)

		server.markScanFailed(
			request,
			scan.ID,
			"Gagal memasukkan scan ke antrean",
		)

		writeError(
			writer,
			http.StatusInternalServerError,
			"queue_error",
			"Gagal memasukkan scan ke antrean",
		)
		return
	}

	writeJSON(
		writer,
		http.StatusCreated,
		toScanResponse(scan),
	)
}

func (server *Server) getScan(
	writer http.ResponseWriter,
	request *http.Request,
) {
	scanID, err := uuid.Parse(chi.URLParam(request, "scanId"))
	if err != nil {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_scan_id",
			"Format scan ID tidak valid",
		)
		return
	}

	scan, err := server.queries.GetScan(
		request.Context(),
		scanID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(
				writer,
				http.StatusNotFound,
				"scan_not_found",
				"Data scan tidak ditemukan",
			)
			return
		}

		log.Printf("gagal mengambil scan: %v", err)

		writeError(
			writer,
			http.StatusInternalServerError,
			"database_error",
			"Gagal mengambil data scan",
		)
		return
	}

	writeJSON(
		writer,
		http.StatusOK,
		toScanResponse(scan),
	)
}

func (server *Server) markScanFailed(
	request *http.Request,
	scanID uuid.UUID,
	message string,
) {
	_, err := server.queries.UpdateScanStatus(
		request.Context(),
		db.UpdateScanStatusParams{
			ID:           scanID,
			Status:       db.ScanStatus("failed"),
			ErrorMessage: message,
		},
	)
	if err != nil {
		log.Printf("gagal mengubah status scan menjadi failed: %v", err)
	}
}

func (server *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			origin := request.Header.Get("Origin")

			if origin != "" && origin == server.cfg.WebOrigin {
				writer.Header().Set(
					"Access-Control-Allow-Origin",
					origin,
				)
				writer.Header().Set(
					"Access-Control-Allow-Credentials",
					"true",
				)
				writer.Header().Add("Vary", "Origin")
			}

			writer.Header().Set(
				"Access-Control-Allow-Headers",
				"Accept, Authorization, Content-Type, X-CSRF-Token",
			)
			writer.Header().Set(
				"Access-Control-Allow-Methods",
				"GET, POST, PATCH, DELETE, OPTIONS",
			)

			if request.Method == http.MethodOptions {
				writer.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(writer, request)
		},
	)
}

func readJSON(
	writer http.ResponseWriter,
	request *http.Request,
	destination any,
) error {
	request.Body = http.MaxBytesReader(
		writer,
		request.Body,
		1<<20,
	)

	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destination); err != nil {
		return err
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("body hanya boleh berisi satu objek JSON")
	}

	return nil
}

func toScanResponse(scan db.Scan) scanResponse {
	return scanResponse{
		ID:           scan.ID.String(),
		URL:          scan.URL,
		Status:       string(scan.Status),
		ErrorMessage: scan.ErrorMessage,
		CreatedAt:    scan.CreatedAt,
		UpdatedAt:    scan.UpdatedAt,
	}
}

func writeError(
	writer http.ResponseWriter,
	statusCode int,
	code string,
	message string,
) {
	writeJSON(
		writer,
		statusCode,
		errorResponse{
			Code:    code,
			Message: message,
		},
	)
}

func writeJSON(
	writer http.ResponseWriter,
	statusCode int,
	payload any,
) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)

	_ = json.NewEncoder(writer).Encode(payload)
}
