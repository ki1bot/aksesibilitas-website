package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/ki1bot/aksesibilitas-website/internal/auth"
	"github.com/ki1bot/aksesibilitas-website/internal/config"
	db "github.com/ki1bot/aksesibilitas-website/internal/database/db"
	taskqueue "github.com/ki1bot/aksesibilitas-website/internal/queue"
	reportbuilder "github.com/ki1bot/aksesibilitas-website/internal/report"
	"github.com/ki1bot/aksesibilitas-website/internal/security"
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

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type projectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type scanRequest struct {
	ProjectID string `json:"project_id"`
	URL       string `json:"url"`
}

type reviewRequest struct {
	Status string `json:"status"`
	Notes  string `json:"notes"`
}

type reportRequest struct {
	Format string `json:"format"`
}

type authResponse struct {
	User      db.User `json:"user"`
	CSRFToken string  `json:"csrf_token"`
}

type healthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Redis    string `json:"redis"`
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type manualReviewResponse struct {
	Review db.ManualReview       `json:"review"`
	Items  []db.ManualReviewItem `json:"items"`
}

type reportResponse struct {
	ID          uuid.UUID       `json:"id"`
	ScanID      uuid.UUID       `json:"scan_id"`
	Format      db.ReportFormat `json:"format"`
	Filename    string          `json:"filename"`
	ContentType string          `json:"content_type"`
	CreatedAt   time.Time       `json:"created_at"`
}

type violationResponse struct {
	Violation db.Violation       `json:"violation"`
	Nodes     []db.ViolationNode `json:"nodes"`
}

var manualReviewTemplates = []struct {
	Criterion   string
	Instruction string
}{
	{
		Criterion:   "Navigasi hanya dengan keyboard",
		Instruction: "Gunakan Tab, Shift+Tab, Enter, Space, dan tombol panah. Pastikan seluruh kontrol dapat digunakan tanpa mouse.",
	},
	{
		Criterion:   "Indikator fokus terlihat",
		Instruction: "Pastikan komponen interaktif menampilkan indikator fokus yang jelas dan tidak tertutup elemen lain.",
	},
	{
		Criterion:   "Urutan fokus masuk akal",
		Instruction: "Periksa bahwa perpindahan fokus mengikuti urutan visual dan urutan membaca halaman.",
	},
	{
		Criterion:   "Pembesaran halaman hingga 200 persen",
		Instruction: "Perbesar halaman hingga 200 persen dan pastikan konten tetap dapat dibaca serta digunakan.",
	},
	{
		Criterion:   "Pengujian pembaca layar",
		Instruction: "Periksa nama, peran, status, heading, landmark, dan pesan kesalahan menggunakan pembaca layar.",
	},
	{
		Criterion:   "Makna tidak bergantung pada warna",
		Instruction: "Pastikan informasi penting tidak disampaikan hanya melalui perbedaan warna.",
	},
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

	router.Route(
		"/api/v1",
		func(router chi.Router) {
			router.Get("/health", server.health)

			router.Route(
				"/auth",
				func(router chi.Router) {
					router.Post(
						"/register",
						server.register,
					)

					router.Post(
						"/login",
						server.login,
					)
				},
			)

			router.Group(
				func(router chi.Router) {
					router.Use(server.authenticate)
					router.Use(server.csrf)

					router.Get(
						"/auth/me",
						server.me,
					)

					router.Patch(
						"/auth/me",
						server.updateProfile,
					)

					router.Post(
						"/auth/logout",
						server.logout,
					)

					router.Get(
						"/projects",
						server.listProjects,
					)

					router.Post(
						"/projects",
						server.createProject,
					)

					router.Route(
						"/projects/{projectId}",
						func(router chi.Router) {
							router.Get(
								"/",
								server.getProject,
							)

							router.Patch(
								"/",
								server.updateProject,
							)

							router.Delete(
								"/",
								server.deleteProject,
							)
						},
					)

					router.Get(
						"/scans",
						server.listScans,
					)

					router.Post(
						"/scans",
						server.createScan,
					)

					router.Route(
						"/scans/{scanId}",
						func(router chi.Router) {
							router.Get(
								"/",
								server.getScan,
							)

							router.Delete(
								"/",
								server.deleteScan,
							)

							router.Post(
								"/cancel",
								server.cancelScan,
							)

							router.Post(
								"/retry",
								server.retryScan,
							)

							router.Get(
								"/violations",
								server.listViolations,
							)

							router.Get(
								"/manual-review",
								server.getManualReview,
							)

							router.Post(
								"/reports",
								server.createReport,
							)
						},
					)

					router.Route(
						"/violations/{violationId}",
						func(router chi.Router) {
							router.Get(
								"/",
								server.getViolation,
							)

							router.Patch(
								"/",
								server.updateViolation,
							)
						},
					)

					router.Patch(
						"/manual-review/items/{itemId}",
						server.updateManualReviewItem,
					)

					router.Get(
						"/reports/{reportId}",
						server.getReport,
					)

					router.Get(
						"/reports/{reportId}/download",
						server.downloadReport,
					)
				},
			)
		},
	)

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

	if err := server.pool.Ping(
		request.Context(),
	); err != nil {
		response.Status = "degraded"
		response.Database = "unavailable"
		statusCode = http.StatusServiceUnavailable
	}

	if err := server.redisClient.Ping(
		request.Context(),
	).Err(); err != nil {
		response.Status = "degraded"
		response.Redis = "unavailable"
		statusCode = http.StatusServiceUnavailable
	}

	writeJSON(writer, statusCode, response)
}

func (server *Server) register(
	writer http.ResponseWriter,
	request *http.Request,
) {
	allowed, err := server.allowRequest(
		request.Context(),
		"register:"+clientIP(request),
		5,
		time.Minute,
	)
	if err != nil || !allowed {
		writeError(
			writer,
			http.StatusTooManyRequests,
			"rate_limit_exceeded",
			"Terlalu banyak percobaan pendaftaran",
		)
		return
	}

	var input registerRequest

	if err := readJSON(
		writer,
		request,
		&input,
	); err != nil {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_request",
			"Body JSON tidak valid",
		)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(
		strings.TrimSpace(input.Email),
	)

	if len(input.Name) < 2 ||
		len(input.Name) > 100 {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_name",
			"Nama harus terdiri dari 2 sampai 100 karakter",
		)
		return
	}

	if _, err := mail.ParseAddress(
		input.Email,
	); err != nil {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_email",
			"Alamat email tidak valid",
		)
		return
	}

	passwordHash, err := auth.HashPassword(
		input.Password,
	)
	if err != nil {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_password",
			err.Error(),
		)
		return
	}

	transaction, err := server.pool.Begin(
		request.Context(),
	)
	if err != nil {
		writeInternalError(writer)
		return
	}
	defer transaction.Rollback(request.Context())

	queries := server.queries.WithTx(transaction)

	user, err := queries.CreateUser(
		request.Context(),
		db.CreateUserParams{
			ID:           uuid.New(),
			Name:         input.Name,
			Email:        input.Email,
			PasswordHash: passwordHash,
		},
	)
	if err != nil {
		var databaseError *pgconn.PgError

		if errors.As(err, &databaseError) &&
			databaseError.Code == "23505" {
			writeError(
				writer,
				http.StatusConflict,
				"email_exists",
				"Email sudah terdaftar",
			)
			return
		}

		log.Printf("gagal membuat user: %v", err)
		writeInternalError(writer)
		return
	}

	project, err := queries.CreateProject(
		request.Context(),
		db.CreateProjectParams{
			ID:          uuid.New(),
			OwnerID:     user.ID,
			Name:        "Project pertama",
			Description: "Project awal AksesCheck ID",
		},
	)
	if err != nil {
		writeInternalError(writer)
		return
	}

	if err := queries.AddProjectMember(
		request.Context(),
		db.AddProjectMemberParams{
			ProjectID: project.ID,
			UserID:    user.ID,
			Role:      "owner",
		},
	); err != nil {
		writeInternalError(writer)
		return
	}

	if err := transaction.Commit(
		request.Context(),
	); err != nil {
		writeInternalError(writer)
		return
	}

	tokens, err := server.sessions.Create(
		request.Context(),
		user.ID,
		request.UserAgent(),
		clientIP(request),
	)
	if err != nil {
		writeInternalError(writer)
		return
	}

	server.sessions.SetCookies(
		writer,
		tokens.SessionToken,
		tokens.CSRFToken,
		tokens.ExpiresAt,
	)

	writeJSON(
		writer,
		http.StatusCreated,
		authResponse{
			User:      publicUser(user),
			CSRFToken: tokens.CSRFToken,
		},
	)
}

func (server *Server) login(
	writer http.ResponseWriter,
	request *http.Request,
) {
	allowed, err := server.allowRequest(
		request.Context(),
		"login:"+clientIP(request),
		10,
		time.Minute,
	)
	if err != nil || !allowed {
		writeError(
			writer,
			http.StatusTooManyRequests,
			"rate_limit_exceeded",
			"Terlalu banyak percobaan login",
		)
		return
	}

	var input loginRequest

	if err := readJSON(
		writer,
		request,
		&input,
	); err != nil {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_request",
			"Body JSON tidak valid",
		)
		return
	}

	user, err := server.queries.GetUserByEmail(
		request.Context(),
		strings.TrimSpace(input.Email),
	)
	if err != nil ||
		!auth.VerifyPassword(
			user.PasswordHash,
			input.Password,
		) {
		writeError(
			writer,
			http.StatusUnauthorized,
			"invalid_credentials",
			"Email atau password salah",
		)
		return
	}

	tokens, err := server.sessions.Create(
		request.Context(),
		user.ID,
		request.UserAgent(),
		clientIP(request),
	)
	if err != nil {
		writeInternalError(writer)
		return
	}

	server.sessions.SetCookies(
		writer,
		tokens.SessionToken,
		tokens.CSRFToken,
		tokens.ExpiresAt,
	)

	writeJSON(
		writer,
		http.StatusOK,
		authResponse{
			User:      publicUser(user),
			CSRFToken: tokens.CSRFToken,
		},
	)
}

func (server *Server) logout(
	writer http.ResponseWriter,
	request *http.Request,
) {
	if err := server.sessions.Destroy(
		request.Context(),
		request,
	); err != nil {
		writeInternalError(writer)
		return
	}

	server.sessions.ClearCookie(writer)
	writer.WriteHeader(http.StatusNoContent)
}

func (server *Server) me(
	writer http.ResponseWriter,
	request *http.Request,
) {
	principal := principalFromContext(
		request.Context(),
	)

	writeJSON(
		writer,
		http.StatusOK,
		publicUser(principal.User),
	)
}

func (server *Server) listProjects(
	writer http.ResponseWriter,
	request *http.Request,
) {
	principal := principalFromContext(
		request.Context(),
	)

	projects, err :=
		server.queries.ListProjectsByUser(
			request.Context(),
			principal.User.ID,
		)
	if err != nil {
		writeInternalError(writer)
		return
	}

	writeJSON(writer, http.StatusOK, projects)
}

func (server *Server) createProject(
	writer http.ResponseWriter,
	request *http.Request,
) {
	var input projectRequest

	if err := readJSON(
		writer,
		request,
		&input,
	); err != nil {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_request",
			"Body JSON tidak valid",
		)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(
		input.Description,
	)

	if len(input.Name) < 2 ||
		len(input.Name) > 120 {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_name",
			"Nama project harus terdiri dari 2 sampai 120 karakter",
		)
		return
	}

	if len(input.Description) > 1000 {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_description",
			"Deskripsi maksimal 1000 karakter",
		)
		return
	}

	principal := principalFromContext(
		request.Context(),
	)

	transaction, err := server.pool.Begin(
		request.Context(),
	)
	if err != nil {
		writeInternalError(writer)
		return
	}
	defer transaction.Rollback(request.Context())

	queries := server.queries.WithTx(transaction)

	project, err := queries.CreateProject(
		request.Context(),
		db.CreateProjectParams{
			ID:          uuid.New(),
			OwnerID:     principal.User.ID,
			Name:        input.Name,
			Description: input.Description,
		},
	)
	if err != nil {
		writeInternalError(writer)
		return
	}

	if err := queries.AddProjectMember(
		request.Context(),
		db.AddProjectMemberParams{
			ProjectID: project.ID,
			UserID:    principal.User.ID,
			Role:      "owner",
		},
	); err != nil {
		writeInternalError(writer)
		return
	}

	if err := transaction.Commit(
		request.Context(),
	); err != nil {
		writeInternalError(writer)
		return
	}

	writeJSON(
		writer,
		http.StatusCreated,
		project,
	)
}

func (server *Server) getProject(
	writer http.ResponseWriter,
	request *http.Request,
) {
	projectID, ok := parseUUIDParam(
		writer,
		request,
		"projectId",
	)
	if !ok {
		return
	}

	principal := principalFromContext(
		request.Context(),
	)

	project, err := server.queries.GetProjectForUser(
		request.Context(),
		db.ProjectUserParams{
			ProjectID: projectID,
			UserID:    principal.User.ID,
		},
	)
	if err != nil {
		writeDatabaseLookupError(writer, err)
		return
	}

	writeJSON(writer, http.StatusOK, project)
}

func (server *Server) updateProject(
	writer http.ResponseWriter,
	request *http.Request,
) {
	projectID, ok := parseUUIDParam(
		writer,
		request,
		"projectId",
	)
	if !ok {
		return
	}

	var input projectRequest

	if err := readJSON(
		writer,
		request,
		&input,
	); err != nil {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_request",
			"Body JSON tidak valid",
		)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(
		input.Description,
	)

	if len(input.Name) < 2 ||
		len(input.Name) > 120 ||
		len(input.Description) > 1000 {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_project",
			"Data project tidak valid",
		)
		return
	}

	principal := principalFromContext(
		request.Context(),
	)

	project, err := server.queries.UpdateProject(
		request.Context(),
		db.UpdateProjectParams{
			ProjectID:   projectID,
			UserID:      principal.User.ID,
			Name:        input.Name,
			Description: input.Description,
		},
	)
	if err != nil {
		writeDatabaseLookupError(writer, err)
		return
	}

	writeJSON(writer, http.StatusOK, project)
}

func (server *Server) deleteProject(
	writer http.ResponseWriter,
	request *http.Request,
) {
	projectID, ok := parseUUIDParam(
		writer,
		request,
		"projectId",
	)
	if !ok {
		return
	}

	principal := principalFromContext(
		request.Context(),
	)

	err := server.queries.DeleteProject(
		request.Context(),
		db.ProjectUserParams{
			ProjectID: projectID,
			UserID:    principal.User.ID,
		},
	)
	if err != nil {
		writeDatabaseLookupError(writer, err)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func (server *Server) listScans(
	writer http.ResponseWriter,
	request *http.Request,
) {
	principal := principalFromContext(
		request.Context(),
	)

	resultLimit := int32(50)

	if rawLimit := request.URL.Query().Get(
		"limit",
	); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err == nil && parsed >= 1 && parsed <= 100 {
			resultLimit = int32(parsed)
		}
	}

	projectValue := request.URL.Query().Get(
		"project_id",
	)

	var scans []db.Scan
	var err error

	if projectValue == "" {
		scans, err = server.queries.ListScansByUser(
			request.Context(),
			principal.User.ID,
			resultLimit,
		)
	} else {
		projectID, parseErr :=
			uuid.Parse(projectValue)

		if parseErr != nil {
			writeError(
				writer,
				http.StatusBadRequest,
				"invalid_project_id",
				"Format project ID tidak valid",
			)
			return
		}

		scans, err =
			server.queries.ListScansByProject(
				request.Context(),
				db.ListScansByProjectParams{
					ProjectID:   projectID,
					UserID:      principal.User.ID,
					ResultLimit: resultLimit,
				},
			)
	}

	if err != nil {
		writeInternalError(writer)
		return
	}

	writeJSON(writer, http.StatusOK, scans)
}

func (server *Server) createScan(
	writer http.ResponseWriter,
	request *http.Request,
) {
	var input scanRequest

	if err := readJSON(
		writer,
		request,
		&input,
	); err != nil {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_request",
			"Body JSON tidak valid",
		)
		return
	}

	projectID, err := uuid.Parse(input.ProjectID)
	if err != nil {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_project_id",
			"Format project ID tidak valid",
		)
		return
	}

	principal := principalFromContext(
		request.Context(),
	)

	_, err = server.queries.GetProjectForUser(
		request.Context(),
		db.ProjectUserParams{
			ProjectID: projectID,
			UserID:    principal.User.ID,
		},
	)
	if err != nil {
		writeDatabaseLookupError(writer, err)
		return
	}

	count, err :=
		server.queries.CountRecentScansByUser(
			request.Context(),
			db.CountRecentScansByUserParams{
				UserID: principal.User.ID,
				SinceTime: time.Now().Add(
					-time.Minute,
				),
			},
		)
	if err != nil {
		writeInternalError(writer)
		return
	}

	if count >= 5 {
		writeError(
			writer,
			http.StatusTooManyRequests,
			"scan_rate_limit",
			"Maksimal lima scan per menit",
		)
		return
	}

	normalizedURL, err :=
		security.ValidatePublicHTTPURL(
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

	transaction, err := server.pool.Begin(
		request.Context(),
	)
	if err != nil {
		writeInternalError(writer)
		return
	}
	defer transaction.Rollback(request.Context())

	queries := server.queries.WithTx(transaction)

	scan, err := queries.CreateScan(
		request.Context(),
		db.CreateScanParams{
			ID:        uuid.New(),
			ProjectID: projectID,
			CreatedBy: principal.User.ID,
			URL:       normalizedURL,
		},
	)
	if err != nil {
		writeInternalError(writer)
		return
	}

	manualReview, err :=
		queries.CreateManualReview(
			request.Context(),
			db.CreateManualReviewParams{
				ID:     uuid.New(),
				ScanID: scan.ID,
			},
		)
	if err != nil {
		writeInternalError(writer)
		return
	}

	for index, template := range manualReviewTemplates {
		_, err := queries.CreateManualReviewItem(
			request.Context(),
			db.CreateManualReviewItemParams{
				ID:             uuid.New(),
				ManualReviewID: manualReview.ID,
				Criterion:      template.Criterion,
				Instruction:    template.Instruction,
				Position:       int32(index + 1),
			},
		)
		if err != nil {
			writeInternalError(writer)
			return
		}
	}

	if err := transaction.Commit(
		request.Context(),
	); err != nil {
		writeInternalError(writer)
		return
	}

	task, options, err :=
		taskqueue.NewAccessibilityScanTask(
			scan.ID.String(),
			scan.URL,
			server.cfg.ScanQueue,
			server.cfg.ScanTimeout,
		)
	if err != nil {
		server.failQueuedScan(
			request.Context(),
			scan.ID,
			"Gagal membuat pekerjaan scan",
		)

		writeInternalError(writer)
		return
	}

	if _, err := server.queueClient.Enqueue(
		task,
		options...,
	); err != nil {
		server.failQueuedScan(
			request.Context(),
			scan.ID,
			"Gagal memasukkan scan ke antrean",
		)

		writeInternalError(writer)
		return
	}

	writeJSON(
		writer,
		http.StatusCreated,
		scan,
	)
}

func (server *Server) getScan(
	writer http.ResponseWriter,
	request *http.Request,
) {
	scanID, ok := parseUUIDParam(
		writer,
		request,
		"scanId",
	)
	if !ok {
		return
	}

	principal := principalFromContext(
		request.Context(),
	)

	scan, err := server.queries.GetScanForUser(
		request.Context(),
		db.ScanUserParams{
			ScanID: scanID,
			UserID: principal.User.ID,
		},
	)
	if err != nil {
		writeDatabaseLookupError(writer, err)
		return
	}

	writeJSON(writer, http.StatusOK, scan)
}

func (server *Server) cancelScan(
	writer http.ResponseWriter,
	request *http.Request,
) {
	scanID, ok := parseUUIDParam(
		writer,
		request,
		"scanId",
	)
	if !ok {
		return
	}

	principal := principalFromContext(
		request.Context(),
	)

	scan, err := server.queries.CancelScan(
		request.Context(),
		db.ScanUserParams{
			ScanID: scanID,
			UserID: principal.User.ID,
		},
	)
	if err != nil {
		writeDatabaseLookupError(writer, err)
		return
	}

	writeJSON(writer, http.StatusOK, scan)
}

func (server *Server) retryScan(
	writer http.ResponseWriter,
	request *http.Request,
) {
	scanID, ok := parseUUIDParam(
		writer,
		request,
		"scanId",
	)
	if !ok {
		return
	}

	principal := principalFromContext(
		request.Context(),
	)

	transaction, err := server.pool.Begin(
		request.Context(),
	)
	if err != nil {
		writeInternalError(writer)
		return
	}
	defer transaction.Rollback(request.Context())

	queries := server.queries.WithTx(transaction)

	scan, err := queries.ResetScanForRetry(
		request.Context(),
		db.ScanUserParams{
			ScanID: scanID,
			UserID: principal.User.ID,
		},
	)
	if err != nil {
		writeDatabaseLookupError(writer, err)
		return
	}

	if err := queries.DeleteScanResults(
		request.Context(),
		scanID,
	); err != nil {
		writeInternalError(writer)
		return
	}

	if err := queries.DeleteScanReports(
		request.Context(),
		scanID,
	); err != nil {
		writeInternalError(writer)
		return
	}

	if err := queries.ResetManualReview(
		request.Context(),
		scanID,
	); err != nil {
		writeInternalError(writer)
		return
	}

	if err := transaction.Commit(
		request.Context(),
	); err != nil {
		writeInternalError(writer)
		return
	}

	task, options, err :=
		taskqueue.NewAccessibilityScanTask(
			scan.ID.String(),
			scan.URL,
			server.cfg.ScanQueue,
			server.cfg.ScanTimeout,
		)
	if err != nil {
		server.failQueuedScan(
			request.Context(),
			scan.ID,
			"Gagal membuat pekerjaan retry",
		)

		writeInternalError(writer)
		return
	}

	if _, err := server.queueClient.Enqueue(
		task,
		options...,
	); err != nil {
		server.failQueuedScan(
			request.Context(),
			scan.ID,
			"Gagal memasukkan retry ke antrean",
		)

		writeInternalError(writer)
		return
	}

	writeJSON(writer, http.StatusOK, scan)
}

func (server *Server) deleteScan(
	writer http.ResponseWriter,
	request *http.Request,
) {
	scanID, ok := parseUUIDParam(
		writer,
		request,
		"scanId",
	)
	if !ok {
		return
	}

	principal := principalFromContext(
		request.Context(),
	)

	err := server.queries.DeleteScanForUser(
		request.Context(),
		db.ScanUserParams{
			ScanID: scanID,
			UserID: principal.User.ID,
		},
	)
	if err != nil {
		writeDatabaseLookupError(writer, err)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func (server *Server) listViolations(
	writer http.ResponseWriter,
	request *http.Request,
) {
	scanID, ok := parseUUIDParam(
		writer,
		request,
		"scanId",
	)
	if !ok {
		return
	}

	principal := principalFromContext(
		request.Context(),
	)

	violations, err :=
		server.queries.ListViolationsForScan(
			request.Context(),
			db.ScanUserParams{
				ScanID: scanID,
				UserID: principal.User.ID,
			},
		)
	if err != nil {
		writeInternalError(writer)
		return
	}

	writeJSON(
		writer,
		http.StatusOK,
		violations,
	)
}

func (server *Server) getViolation(
	writer http.ResponseWriter,
	request *http.Request,
) {
	violationID, ok := parseUUIDParam(
		writer,
		request,
		"violationId",
	)
	if !ok {
		return
	}

	principal := principalFromContext(
		request.Context(),
	)

	violation, err :=
		server.queries.GetViolationForUser(
			request.Context(),
			db.ViolationUserParams{
				ViolationID: violationID,
				UserID:      principal.User.ID,
			},
		)
	if err != nil {
		writeDatabaseLookupError(writer, err)
		return
	}

	nodes, err :=
		server.queries.ListViolationNodes(
			request.Context(),
			violation.ID,
		)
	if err != nil {
		writeInternalError(writer)
		return
	}

	writeJSON(
		writer,
		http.StatusOK,
		violationResponse{
			Violation: violation,
			Nodes:     nodes,
		},
	)
}

func (server *Server) updateViolation(
	writer http.ResponseWriter,
	request *http.Request,
) {
	violationID, ok := parseUUIDParam(
		writer,
		request,
		"violationId",
	)
	if !ok {
		return
	}

	var input reviewRequest

	if err := readJSON(
		writer,
		request,
		&input,
	); err != nil {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_request",
			"Body JSON tidak valid",
		)
		return
	}

	status, valid := parseReviewStatus(
		input.Status,
	)
	if !valid || len(input.Notes) > 5000 {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_review",
			"Status atau catatan review tidak valid",
		)
		return
	}

	principal := principalFromContext(
		request.Context(),
	)

	violation, err :=
		server.queries.UpdateViolationReview(
			request.Context(),
			db.UpdateViolationReviewParams{
				ViolationID:  violationID,
				UserID:       principal.User.ID,
				ReviewStatus: status,
				Notes:        strings.TrimSpace(input.Notes),
			},
		)
	if err != nil {
		writeDatabaseLookupError(writer, err)
		return
	}

	writeJSON(writer, http.StatusOK, violation)
}

func (server *Server) getManualReview(
	writer http.ResponseWriter,
	request *http.Request,
) {
	scanID, ok := parseUUIDParam(
		writer,
		request,
		"scanId",
	)
	if !ok {
		return
	}

	principal := principalFromContext(
		request.Context(),
	)

	review, err :=
		server.queries.GetManualReviewForScan(
			request.Context(),
			db.ScanUserParams{
				ScanID: scanID,
				UserID: principal.User.ID,
			},
		)
	if err != nil {
		writeDatabaseLookupError(writer, err)
		return
	}

	items, err :=
		server.queries.ListManualReviewItems(
			request.Context(),
			review.ID,
		)
	if err != nil {
		writeInternalError(writer)
		return
	}

	writeJSON(
		writer,
		http.StatusOK,
		manualReviewResponse{
			Review: review,
			Items:  items,
		},
	)
}

func (server *Server) updateManualReviewItem(
	writer http.ResponseWriter,
	request *http.Request,
) {
	itemID, ok := parseUUIDParam(
		writer,
		request,
		"itemId",
	)
	if !ok {
		return
	}

	var input reviewRequest

	if err := readJSON(
		writer,
		request,
		&input,
	); err != nil {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_request",
			"Body JSON tidak valid",
		)
		return
	}

	status, valid := parseReviewStatus(
		input.Status,
	)
	if !valid || len(input.Notes) > 5000 {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_review",
			"Status atau catatan review tidak valid",
		)
		return
	}

	principal := principalFromContext(
		request.Context(),
	)

	transaction, err := server.pool.Begin(
		request.Context(),
	)
	if err != nil {
		writeInternalError(writer)
		return
	}
	defer transaction.Rollback(request.Context())

	queries := server.queries.WithTx(transaction)

	item, err := queries.UpdateManualReviewItem(
		request.Context(),
		db.UpdateManualReviewItemParams{
			ItemID: itemID,
			UserID: principal.User.ID,
			Status: status,
			Notes:  strings.TrimSpace(input.Notes),
		},
	)
	if err != nil {
		writeDatabaseLookupError(writer, err)
		return
	}

	if err := queries.RefreshManualReviewStatus(
		request.Context(),
		item.ManualReviewID,
	); err != nil {
		writeInternalError(writer)
		return
	}

	if err := transaction.Commit(
		request.Context(),
	); err != nil {
		writeInternalError(writer)
		return
	}

	writeJSON(writer, http.StatusOK, item)
}

func (server *Server) createReport(
	writer http.ResponseWriter,
	request *http.Request,
) {
	scanID, ok := parseUUIDParam(
		writer,
		request,
		"scanId",
	)
	if !ok {
		return
	}

	var input reportRequest

	if err := readJSON(
		writer,
		request,
		&input,
	); err != nil {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_request",
			"Body JSON tidak valid",
		)
		return
	}

	format := db.ReportFormat(
		strings.ToLower(
			strings.TrimSpace(input.Format),
		),
	)

	if format != db.ReportFormatJSON &&
		format != db.ReportFormatPDF {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_report_format",
			"Format laporan harus json atau pdf",
		)
		return
	}

	principal := principalFromContext(
		request.Context(),
	)

	scan, err := server.queries.GetScanForUser(
		request.Context(),
		db.ScanUserParams{
			ScanID: scanID,
			UserID: principal.User.ID,
		},
	)
	if err != nil {
		writeDatabaseLookupError(writer, err)
		return
	}

	if scan.Status != db.ScanStatusCompleted {
		writeError(
			writer,
			http.StatusConflict,
			"scan_not_completed",
			"Laporan hanya dapat dibuat setelah scan selesai",
		)
		return
	}

	violations, err :=
		server.loadViolationsWithNodes(
			request.Context(),
			scanID,
			principal.User.ID,
		)
	if err != nil {
		writeInternalError(writer)
		return
	}

	review, err :=
		server.queries.GetManualReviewForScan(
			request.Context(),
			db.ScanUserParams{
				ScanID: scanID,
				UserID: principal.User.ID,
			},
		)
	if err != nil {
		writeInternalError(writer)
		return
	}

	manualItems, err :=
		server.queries.ListManualReviewItems(
			request.Context(),
			review.ID,
		)
	if err != nil {
		writeInternalError(writer)
		return
	}

	var content []byte
	var filename string
	var contentType string

	switch format {
	case db.ReportFormatPDF:
		content, err = reportbuilder.BuildPDF(
			scan,
			violations,
			manualItems,
		)

		filename = "aksescheck-" +
			scan.ID.String() +
			".pdf"

		contentType = "application/pdf"
	default:
		content, err = reportbuilder.BuildJSON(
			scan,
			violations,
			manualItems,
		)

		filename = "aksescheck-" +
			scan.ID.String() +
			".json"

		contentType = "application/json"
	}

	if err != nil {
		writeInternalError(writer)
		return
	}

	report, err := server.queries.CreateReport(
		request.Context(),
		db.CreateReportParams{
			ID:          uuid.New(),
			ScanID:      scan.ID,
			CreatedBy:   principal.User.ID,
			Format:      format,
			Filename:    filename,
			ContentType: contentType,
			Content:     content,
		},
	)
	if err != nil {
		writeInternalError(writer)
		return
	}

	writeJSON(
		writer,
		http.StatusCreated,
		toReportResponse(report),
	)
}

func (server *Server) getReport(
	writer http.ResponseWriter,
	request *http.Request,
) {
	report, ok := server.findReport(
		writer,
		request,
	)
	if !ok {
		return
	}

	writeJSON(
		writer,
		http.StatusOK,
		toReportResponse(report),
	)
}

func (server *Server) downloadReport(
	writer http.ResponseWriter,
	request *http.Request,
) {
	report, ok := server.findReport(
		writer,
		request,
	)
	if !ok {
		return
	}

	writer.Header().Set(
		"Content-Type",
		report.ContentType,
	)

	writer.Header().Set(
		"Content-Disposition",
		`attachment; filename="`+
			strings.ReplaceAll(
				report.Filename,
				`"`,
				"",
			)+
			`"`,
	)

	writer.Header().Set(
		"Content-Length",
		strconv.Itoa(len(report.Content)),
	)

	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(report.Content)
}

func (server *Server) findReport(
	writer http.ResponseWriter,
	request *http.Request,
) (db.Report, bool) {
	reportID, ok := parseUUIDParam(
		writer,
		request,
		"reportId",
	)
	if !ok {
		return db.Report{}, false
	}

	principal := principalFromContext(
		request.Context(),
	)

	report, err :=
		server.queries.GetReportForUser(
			request.Context(),
			db.ReportUserParams{
				ReportID: reportID,
				UserID:   principal.User.ID,
			},
		)
	if err != nil {
		writeDatabaseLookupError(writer, err)
		return db.Report{}, false
	}

	return report, true
}

func (server *Server) loadViolationsWithNodes(
	ctx context.Context,
	scanID uuid.UUID,
	userID uuid.UUID,
) ([]db.ViolationWithNodes, error) {
	violations, err :=
		server.queries.ListViolationsForScan(
			ctx,
			db.ScanUserParams{
				ScanID: scanID,
				UserID: userID,
			},
		)
	if err != nil {
		return nil, err
	}

	result := make(
		[]db.ViolationWithNodes,
		0,
		len(violations),
	)

	for _, violation := range violations {
		nodes, nodeErr :=
			server.queries.ListViolationNodes(
				ctx,
				violation.ID,
			)
		if nodeErr != nil {
			return nil, nodeErr
		}

		result = append(
			result,
			db.ViolationWithNodes{
				Violation: violation,
				Nodes:     nodes,
			},
		)
	}

	return result, nil
}

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

func (server *Server) failQueuedScan(
	ctx context.Context,
	scanID uuid.UUID,
	message string,
) {
	_, err := server.queries.FailScan(
		ctx,
		scanID,
		message,
	)
	if err != nil {
		log.Printf(
			"gagal mengubah status scan %s: %v",
			scanID,
			err,
		)
	}
}

func parseUUIDParam(
	writer http.ResponseWriter,
	request *http.Request,
	name string,
) (uuid.UUID, bool) {
	value := chi.URLParam(request, name)

	id, err := uuid.Parse(value)
	if err != nil {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_id",
			"Format ID tidak valid",
		)
		return uuid.Nil, false
	}

	return id, true
}

func parseReviewStatus(
	value string,
) (db.ReviewStatus, bool) {
	status := db.ReviewStatus(
		strings.ToLower(
			strings.TrimSpace(value),
		),
	)

	switch status {
	case db.ReviewStatusPending,
		db.ReviewStatusPassed,
		db.ReviewStatusFailed,
		db.ReviewStatusNotApplicable:
		return status, true
	default:
		return "", false
	}
}

func principalFromContext(
	ctx context.Context,
) auth.Principal {
	principal, _ := ctx.Value(
		principalContextKey{},
	).(auth.Principal)

	return principal
}

func clientIP(request *http.Request) string {
	host, _, err := net.SplitHostPort(
		request.RemoteAddr,
	)
	if err == nil {
		return host
	}

	return request.RemoteAddr
}

func publicUser(user db.User) db.User {
	user.PasswordHash = ""
	return user
}

func toReportResponse(
	report db.Report,
) reportResponse {
	return reportResponse{
		ID:          report.ID,
		ScanID:      report.ScanID,
		Format:      report.Format,
		Filename:    report.Filename,
		ContentType: report.ContentType,
		CreatedAt:   report.CreatedAt,
	}
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

	if err := decoder.Decode(
		&struct{}{},
	); !errors.Is(err, io.EOF) {
		return errors.New(
			"body hanya boleh berisi satu objek JSON",
		)
	}

	return nil
}

func writeDatabaseLookupError(
	writer http.ResponseWriter,
	err error,
) {
	if errors.Is(err, pgx.ErrNoRows) ||
		errors.Is(err, db.ErrNotFound) {
		writeError(
			writer,
			http.StatusNotFound,
			"not_found",
			"Data tidak ditemukan",
		)
		return
	}

	log.Printf("kesalahan database: %v", err)
	writeInternalError(writer)
}

func writeInternalError(
	writer http.ResponseWriter,
) {
	writeError(
		writer,
		http.StatusInternalServerError,
		"internal_error",
		"Terjadi kesalahan pada server",
	)
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
	writer.Header().Set(
		"Content-Type",
		"application/json; charset=utf-8",
	)

	writer.WriteHeader(statusCode)

	if payload != nil {
		_ = json.NewEncoder(writer).Encode(
			payload,
		)
	}
}
