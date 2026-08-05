package account

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/ki1bot/aksesibilitas-website/internal/auth"
	"github.com/ki1bot/aksesibilitas-website/internal/config"
)

type Handler struct {
	cfg   config.Config
	pool  *pgxpool.Pool
	redis *redis.Client
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	Token                string `json:"token"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
}

type changePasswordRequest struct {
	CurrentPassword      string `json:"current_password"`
	NewPassword          string `json:"new_password"`
	PasswordConfirmation string `json:"password_confirmation"`
}

type forgotPasswordResponse struct {
	Message       string `json:"message"`
	DebugResetURL string `json:"debug_reset_url,omitempty"`
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type resetUser struct {
	ID    uuid.UUID
	Name  string
	Email string
}

type authenticatedUser struct {
	ID           uuid.UUID
	PasswordHash string
}

func NewHandler(
	cfg config.Config,
	pool *pgxpool.Pool,
	redisClient *redis.Client,
) *Handler {
	return &Handler{
		cfg:   cfg,
		pool:  pool,
		redis: redisClient,
	}
}

func (handler *Handler) Options(
	writer http.ResponseWriter,
	request *http.Request,
) {
	handler.prepareResponse(writer)
	writer.WriteHeader(http.StatusNoContent)
}

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

	sessionHashes, err :=
		handler.replacePasswordAndInvalidateSessions(
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

	handler.removeSessionCache(
		request.Context(),
		sessionHashes,
	)

	handler.clearCookies(writer)
	writer.WriteHeader(http.StatusNoContent)
}

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

func (handler *Handler) authenticate(
	ctx context.Context,
	request *http.Request,
) (authenticatedUser, error) {
	cookie, err := request.Cookie(
		handler.cfg.SessionCookieName,
	)
	if err != nil ||
		strings.TrimSpace(cookie.Value) == "" {
		return authenticatedUser{}, pgx.ErrNoRows
	}

	var user authenticatedUser

	err = handler.pool.QueryRow(
		ctx,
		`
			SELECT users.id, users.password_hash
			FROM sessions
			JOIN users ON users.id = sessions.user_id
			WHERE sessions.token_hash = $1
			  AND sessions.expires_at > NOW()
			LIMIT 1
		`,
		hashToken(cookie.Value),
	).Scan(
		&user.ID,
		&user.PasswordHash,
	)

	return user, err
}

func (handler *Handler) validCSRF(
	request *http.Request,
) bool {
	headerToken := strings.TrimSpace(
		request.Header.Get("X-CSRF-Token"),
	)
	if headerToken == "" {
		return false
	}

	csrfCookie, err := request.Cookie(
		handler.cfg.SessionCookieName + "_csrf",
	)
	if err != nil || csrfCookie.Value == "" {
		return false
	}

	if subtle.ConstantTimeCompare(
		[]byte(headerToken),
		[]byte(csrfCookie.Value),
	) != 1 {
		return false
	}

	sessionCookie, err := request.Cookie(
		handler.cfg.SessionCookieName,
	)
	if err != nil || sessionCookie.Value == "" {
		return false
	}

	var csrfHash string

	err = handler.pool.QueryRow(
		request.Context(),
		`
			SELECT csrf_hash
			FROM sessions
			WHERE token_hash = $1
			  AND expires_at > NOW()
			LIMIT 1
		`,
		hashToken(sessionCookie.Value),
	).Scan(&csrfHash)

	if err != nil {
		return false
	}

	expected, err := hex.DecodeString(csrfHash)
	if err != nil {
		return false
	}

	actual := sha256.Sum256(
		[]byte(headerToken),
	)

	return subtle.ConstantTimeCompare(
		expected,
		actual[:],
	) == 1
}

func (handler *Handler) replacePasswordAndInvalidateSessions(
	ctx context.Context,
	userID uuid.UUID,
	passwordHash string,
) ([]string, error) {
	transaction, err := handler.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer transaction.Rollback(ctx)

	rows, err := transaction.Query(
		ctx,
		`
			SELECT token_hash
			FROM sessions
			WHERE user_id = $1
		`,
		userID,
	)
	if err != nil {
		return nil, err
	}

	sessionHashes := make([]string, 0)

	for rows.Next() {
		var tokenHash string

		if err := rows.Scan(
			&tokenHash,
		); err != nil {
			rows.Close()
			return nil, err
		}

		sessionHashes = append(
			sessionHashes,
			tokenHash,
		)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}

	rows.Close()

	commandTag, err := transaction.Exec(
		ctx,
		`
			UPDATE users
			SET password_hash = $1,
			    updated_at = NOW()
			WHERE id = $2
		`,
		passwordHash,
		userID,
	)
	if err != nil {
		return nil, err
	}

	if commandTag.RowsAffected() != 1 {
		return nil, pgx.ErrNoRows
	}

	if _, err := transaction.Exec(
		ctx,
		`
			DELETE FROM sessions
			WHERE user_id = $1
		`,
		userID,
	); err != nil {
		return nil, err
	}

	if err := transaction.Commit(ctx); err != nil {
		return nil, err
	}

	return sessionHashes, nil
}

func (handler *Handler) removeSessionCache(
	ctx context.Context,
	sessionHashes []string,
) {
	if len(sessionHashes) == 0 {
		return
	}

	keys := make(
		[]string,
		0,
		len(sessionHashes),
	)

	for _, tokenHash := range sessionHashes {
		keys = append(
			keys,
			"session:"+tokenHash,
		)
	}

	if err := handler.redis.Del(
		ctx,
		keys...,
	).Err(); err != nil {
		log.Printf(
			"gagal menghapus cache sesi: %v",
			err,
		)
	}
}

func (handler *Handler) storeResetToken(
	ctx context.Context,
	userID uuid.UUID,
	tokenHash string,
) error {
	userKey := handler.resetUserKey(userID)

	previousHash, err := handler.redis.Get(
		ctx,
		userKey,
	).Result()

	if err != nil &&
		!errors.Is(err, redis.Nil) {
		return err
	}

	pipeline := handler.redis.TxPipeline()

	if previousHash != "" {
		pipeline.Del(
			ctx,
			handler.resetTokenKey(previousHash),
		)
	}

	pipeline.Set(
		ctx,
		handler.resetTokenKey(tokenHash),
		userID.String(),
		handler.cfg.PasswordResetTTL,
	)

	pipeline.Set(
		ctx,
		userKey,
		tokenHash,
		handler.cfg.PasswordResetTTL,
	)

	_, err = pipeline.Exec(ctx)
	return err
}

func (handler *Handler) allowRequest(
	ctx context.Context,
	key string,
	limit int64,
	window time.Duration,
) (bool, error) {
	hash := sha256.Sum256([]byte(key))

	redisKey := "rate-limit:" +
		hex.EncodeToString(hash[:])

	count, err := handler.redis.Incr(
		ctx,
		redisKey,
	).Result()

	if err != nil {
		return false, err
	}

	if count == 1 {
		if err := handler.redis.Expire(
			ctx,
			redisKey,
			window,
		).Err(); err != nil {
			return false, err
		}
	}

	return count <= limit, nil
}

func (handler *Handler) emailConfigured() bool {
	return handler.cfg.SMTPHost != "" &&
		handler.cfg.SMTPPort > 0 &&
		handler.cfg.SMTPFromEmail != ""
}

func (handler *Handler) sendResetEmail(
	user resetUser,
	resetURL string,
) error {
	address := net.JoinHostPort(
		handler.cfg.SMTPHost,
		strconv.Itoa(handler.cfg.SMTPPort),
	)

	var smtpAuth smtp.Auth

	if handler.cfg.SMTPUsername != "" {
		smtpAuth = smtp.PlainAuth(
			"",
			handler.cfg.SMTPUsername,
			handler.cfg.SMTPPassword,
			handler.cfg.SMTPHost,
		)
	}

	subject :=
		"Buat password baru untuk akun AksesCheck"

	body := fmt.Sprintf(
		`<!doctype html>
<html lang="id">
  <body style="font-family:Arial,sans-serif;color:#1c1815;line-height:1.6">
    <h1 style="font-size:22px">Buat password baru</h1>
    <p>Halo %s,</p>
    <p>Kami menerima permintaan untuk mengganti password akun AksesCheck Anda.</p>
    <p><a href="%s" style="display:inline-block;padding:12px 18px;border-radius:10px;background:#0f766e;color:#ffffff;text-decoration:none;font-weight:700">Buat password baru</a></p>
    <p>Tautan ini berlaku selama %s dan hanya dapat digunakan satu kali.</p>
    <p>Jika Anda tidak meminta perubahan ini, abaikan email ini.</p>
  </body>
</html>`,
		html.EscapeString(user.Name),
		html.EscapeString(resetURL),
		humanDuration(
			handler.cfg.PasswordResetTTL,
		),
	)

	message := strings.Join(
		[]string{
			"From: " +
				handler.cfg.SMTPFromName +
				" <" +
				handler.cfg.SMTPFromEmail +
				">",
			"To: " + user.Email,
			"Subject: " + subject,
			"MIME-Version: 1.0",
			"Content-Type: text/html; charset=UTF-8",
			"",
			body,
		},
		"\r\n",
	)

	return smtp.SendMail(
		address,
		smtpAuth,
		handler.cfg.SMTPFromEmail,
		[]string{user.Email},
		[]byte(message),
	)
}

func (handler *Handler) resetTokenKey(
	tokenHash string,
) string {
	return "password-reset:" + tokenHash
}

func (handler *Handler) resetUserKey(
	userID uuid.UUID,
) string {
	return "password-reset-user:" +
		userID.String()
}

func (handler *Handler) clearCookies(
	writer http.ResponseWriter,
) {
	secure := handler.cfg.AppEnv == "production"

	for _, cookie := range []*http.Cookie{
		{
			Name:  handler.cfg.SessionCookieName,
			Value: "",
			Path:  "/",
			Expires: time.Unix(
				0,
				0,
			),
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   secure,
			SameSite: http.SameSiteLaxMode,
		},
		{
			Name: handler.cfg.SessionCookieName +
				"_csrf",
			Value: "",
			Path:  "/",
			Expires: time.Unix(
				0,
				0,
			),
			MaxAge:   -1,
			HttpOnly: false,
			Secure:   secure,
			SameSite: http.SameSiteLaxMode,
		},
	} {
		http.SetCookie(writer, cookie)
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

func randomToken(
	size int,
) (string, error) {
	value := make([]byte, size)

	if _, err := rand.Read(value); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(
		value,
	), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))

	return hex.EncodeToString(hash[:])
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

func humanDuration(
	duration time.Duration,
) string {
	if duration%time.Hour == 0 {
		hours := int(duration / time.Hour)

		return strconv.Itoa(hours) + " jam"
	}

	minutes := int(duration / time.Minute)

	return strconv.Itoa(minutes) + " menit"
}
