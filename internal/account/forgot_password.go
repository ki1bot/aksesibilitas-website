package account

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func (handler *Handler) ForgotPassword(
	writer http.ResponseWriter,
	request *http.Request,
) {
	handler.prepareResponse(writer)

	var input forgotPasswordRequest
	if err := readJSON(writer, request, &input); err != nil {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_request",
			"Data yang dikirim belum benar",
		)
		return
	}

	emailAddress := strings.TrimSpace(input.Email)
	parsedAddress, err := mail.ParseAddress(emailAddress)
	if err != nil {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_email",
			"Alamat email tidak valid",
		)
		return
	}

	emailAddress = strings.ToLower(parsedAddress.Address)

	allowed, err := handler.allowRequest(
		request.Context(),
		"forgot-password:"+clientIP(request)+":"+emailAddress,
		5,
		15*time.Minute,
	)
	if err != nil || !allowed {
		writeError(
			writer,
			http.StatusTooManyRequests,
			"rate_limit_exceeded",
			"Terlalu banyak permintaan. Coba lagi beberapa menit lagi",
		)
		return
	}

	response := forgotPasswordResponse{
		Message: "Jika email tersebut terdaftar, kami akan mengirim tautan untuk membuat password baru.",
	}

	user, err := handler.findUserByEmail(
		request.Context(),
		emailAddress,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		writeJSON(writer, http.StatusAccepted, response)
		return
	}

	if err != nil {
		log.Printf(
			"gagal mencari akun untuk reset password: %v",
			err,
		)
		writeJSON(writer, http.StatusAccepted, response)
		return
	}

	token, err := randomToken(32)
	if err != nil {
		log.Printf(
			"gagal membuat token reset password: %v",
			err,
		)
		writeJSON(writer, http.StatusAccepted, response)
		return
	}

	tokenHash := hashToken(token)

	if err := handler.storeResetToken(
		request.Context(),
		user.ID,
		tokenHash,
	); err != nil {
		log.Printf(
			"gagal menyimpan token reset password: %v",
			err,
		)
		writeJSON(writer, http.StatusAccepted, response)
		return
	}

	resetURL := strings.TrimRight(
		handler.cfg.WebOrigin,
		"/",
	) + "/reset-password?token=" + url.QueryEscape(token)

	if handler.emailConfigured() {
		if err := handler.sendResetEmail(
			user,
			resetURL,
		); err != nil {
			log.Printf(
				"gagal mengirim email reset password: %v",
				err,
			)
		}
	} else if handler.cfg.AppEnv == "production" {
		log.Print(
			"SMTP belum dikonfigurasi. Email reset password tidak dapat dikirim",
		)
	}

	if handler.cfg.AppEnv != "production" {
		response.DebugResetURL = resetURL

		log.Printf(
			"tautan reset password untuk %s: %s",
			user.Email,
			resetURL,
		)
	}

	writeJSON(writer, http.StatusAccepted, response)
}

func (handler *Handler) findUserByEmail(
	ctx context.Context,
	emailAddress string,
) (resetUser, error) {
	var user resetUser

	err := handler.pool.QueryRow(
		ctx,
		`
			SELECT id, name, email
			FROM users
			WHERE LOWER(email) = LOWER($1)
			LIMIT 1
		`,
		emailAddress,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
	)

	return user, err
}
