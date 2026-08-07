package account

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ki1bot/aksesibilitas-website/internal/config"
)

type Handler struct {
	cfg  config.Config
	pool *pgxpool.Pool
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
) *Handler {
	return &Handler{
		cfg:  cfg,
		pool: pool,
	}
}

func (handler *Handler) Options(
	writer http.ResponseWriter,
	request *http.Request,
) {
	handler.prepareResponse(writer)
	writer.WriteHeader(http.StatusNoContent)
}
