package httpapi

import (
	"errors"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/ki1bot/aksesibilitas-website/internal/auth"
	db "github.com/ki1bot/aksesibilitas-website/internal/database/db"
)

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
