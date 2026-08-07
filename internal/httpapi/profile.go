package httpapi

import (
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/ki1bot/aksesibilitas-website/internal/auth"
	db "github.com/ki1bot/aksesibilitas-website/internal/database/db"
)

type updateProfileRequest struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	CurrentPassword string `json:"current_password"`
}

func (server *Server) updateProfile(
	writer http.ResponseWriter,
	request *http.Request,
) {
	var input updateProfileRequest

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
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	nameLength := utf8.RuneCountInString(input.Name)

	if nameLength < 2 || nameLength > 100 {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_name",
			"Nama harus terdiri dari 2 sampai 100 karakter",
		)
		return
	}

	emailLength := utf8.RuneCountInString(input.Email)

	if emailLength < 3 || emailLength > 255 {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_email",
			"Alamat email tidak valid",
		)
		return
	}

	parsedEmail, err := mail.ParseAddress(input.Email)
	if err != nil ||
		!strings.EqualFold(parsedEmail.Address, input.Email) {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_email",
			"Alamat email tidak valid",
		)
		return
	}

	principal := principalFromContext(request.Context())

	emailChanged := !strings.EqualFold(
		input.Email,
		principal.User.Email,
	)

	if emailChanged {
		if input.CurrentPassword == "" {
			writeError(
				writer,
				http.StatusBadRequest,
				"current_password_required",
				"Password saat ini wajib diisi untuk mengganti email",
			)
			return
		}

		if !auth.VerifyPassword(
			principal.User.PasswordHash,
			input.CurrentPassword,
		) {
			writeError(
				writer,
				http.StatusUnauthorized,
				"invalid_credentials",
				"Password saat ini salah",
			)
			return
		}
	}

	user, err := server.queries.UpdateUserProfile(
		request.Context(),
		db.UpdateUserProfileParams{
			ID:    principal.User.ID,
			Name:  input.Name,
			Email: input.Email,
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
				"Email sudah digunakan akun lain",
			)
			return
		}

		writeInternalError(writer)
		return
	}

	writeJSON(
		writer,
		http.StatusOK,
		publicUser(user),
	)
}
