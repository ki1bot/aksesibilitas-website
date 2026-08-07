package httpapi

import (
	"time"

	"github.com/google/uuid"

	db "github.com/ki1bot/aksesibilitas-website/internal/database/db"
)

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
