package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/ki1bot/aksesibilitas-website/internal/auth"
	db "github.com/ki1bot/aksesibilitas-website/internal/database/db"
)

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
