package account

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
)

func (handler *Handler) prepareResponse(
	writer http.ResponseWriter,
) {
	writer.Header().Set(
		"Access-Control-Allow-Origin",
		handler.cfg.WebOrigin,
	)
	writer.Header().Set(
		"Access-Control-Allow-Credentials",
		"true",
	)
	writer.Header().Set(
		"Access-Control-Allow-Methods",
		"POST, OPTIONS",
	)
	writer.Header().Set(
		"Access-Control-Allow-Headers",
		"Content-Type, X-CSRF-Token",
	)
	writer.Header().Set(
		"Cache-Control",
		"no-store",
	)
	writer.Header().Set(
		"X-Content-Type-Options",
		"nosniff",
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

	if err := decoder.Decode(
		destination,
	); err != nil {
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

	_ = json.NewEncoder(writer).Encode(
		payload,
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

func writeInternalError(
	writer http.ResponseWriter,
) {
	writeError(
		writer,
		http.StatusInternalServerError,
		"internal_error",
		"Terjadi masalah pada server. Silakan coba lagi",
	)
}

func clientIP(
	request *http.Request,
) string {
	if forwarded := strings.TrimSpace(
		request.Header.Get("X-Forwarded-For"),
	); forwarded != "" {
		if first, _, found := strings.Cut(
			forwarded,
			",",
		); found {
			return strings.TrimSpace(first)
		}

		return forwarded
	}

	if realIP := strings.TrimSpace(
		request.Header.Get("X-Real-IP"),
	); realIP != "" {
		return realIP
	}

	host, _, err := net.SplitHostPort(
		request.RemoteAddr,
	)
	if err == nil {
		return host
	}

	return request.RemoteAddr
}
