package account

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/ki1bot/aksesibilitas-website/internal/auth"
)

func (handler *Handler) ResetPassword(
	writer http.ResponseWriter,
	request *http.Request,
) {
	handler.prepareResponse(writer)

	allowed, err := handler.allowRequest(
		request.Context(),
		"reset-password:"+clientIP(request),
		10,
		15*time.Minute,
	)
	if err != nil || !allowed {
		writeError(
			writer,
			http.StatusTooManyRequests,
			"rate_limit_exceeded",
			"Terlalu banyak percobaan. Coba lagi beberapa menit lagi",
		)
		return
	}

	var input resetPasswordRequest

	if err := readJSON(
		writer,
		request,
		&input,
	); err != nil {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_request",
			"Data yang dikirim belum benar",
		)
		return
	}

	input.Token = strings.TrimSpace(input.Token)

	if len(input.Token) < 32 {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_token",
			"Tautan reset password tidak valid atau sudah kedaluwarsa",
		)
		return
	}

	if input.Password != input.PasswordConfirmation {
		writeError(
			writer,
			http.StatusBadRequest,
			"password_mismatch",
			"Konfirmasi password belum sama",
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

	tokenHash := hashToken(input.Token)

	userIDValue, err := handler.redis.GetDel(
		request.Context(),
		handler.resetTokenKey(tokenHash),
	).Result()

	if errors.Is(err, redis.Nil) {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_token",
			"Tautan reset password tidak valid atau sudah kedaluwarsa",
		)
		return
	}

	if err != nil {
		writeInternalError(writer)
		return
	}

	userID, err := uuid.Parse(userIDValue)
	if err != nil {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_token",
			"Tautan reset password tidak valid atau sudah kedaluwarsa",
		)
		return
	}

	sessionHashes, err :=
		handler.replacePasswordAndInvalidateSessions(
			request.Context(),
			userID,
			passwordHash,
		)

	if err != nil {
		log.Printf(
			"gagal mengganti password dari token reset: %v",
			err,
		)
		writeInternalError(writer)
		return
	}

	_ = handler.redis.Del(
		request.Context(),
		handler.resetUserKey(userID),
	).Err()

	handler.removeSessionCache(
		request.Context(),
		sessionHashes,
	)

	handler.clearCookies(writer)
	writer.WriteHeader(http.StatusNoContent)
}
