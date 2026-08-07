package account

import (
	"log"
	"net/http"
	"time"

	"github.com/ki1bot/aksesibilitas-website/internal/auth"
)

func (handler *Handler) ChangePassword(
	writer http.ResponseWriter,
	request *http.Request,
) {
	handler.prepareResponse(writer)

	user, err := handler.authenticate(
		request.Context(),
		request,
	)
	if err != nil {
		writeError(
			writer,
			http.StatusUnauthorized,
			"unauthorized",
			"Sesi Anda sudah berakhir. Silakan masuk kembali",
		)
		return
	}

	if !handler.validCSRF(request) {
		writeError(
			writer,
			http.StatusForbidden,
			"invalid_csrf",
			"Permintaan tidak dapat dipastikan keamanannya. Muat ulang halaman lalu coba lagi",
		)
		return
	}

	allowed, err := handler.allowRequest(
		request.Context(),
		"change-password:"+user.ID.String(),
		8,
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

	var input changePasswordRequest

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

	if !auth.VerifyPassword(
		user.PasswordHash,
		input.CurrentPassword,
	) {
		writeError(
			writer,
			http.StatusBadRequest,
			"incorrect_password",
			"Password saat ini salah",
		)
		return
	}

	if input.NewPassword != input.PasswordConfirmation {
		writeError(
			writer,
			http.StatusBadRequest,
			"password_mismatch",
			"Konfirmasi password belum sama",
		)
		return
	}

	if auth.VerifyPassword(
		user.PasswordHash,
		input.NewPassword,
	) {
		writeError(
			writer,
			http.StatusBadRequest,
			"same_password",
			"Password baru harus berbeda dari password saat ini",
		)
		return
	}

	passwordHash, err := auth.HashPassword(
		input.NewPassword,
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

	err = handler.replacePasswordAndInvalidateSessions(
		request.Context(),
		user.ID,
		passwordHash,
	)
	if err != nil {
		log.Printf(
			"gagal mengganti password pengguna: %v",
			err,
		)

		writeInternalError(writer)
		return
	}

	handler.clearCookies(writer)
	writer.WriteHeader(http.StatusNoContent)
}
