package db

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ScanStatus string

const (
	ScanStatusQueued    ScanStatus = "queued"
	ScanStatusRunning   ScanStatus = "running"
	ScanStatusCompleted ScanStatus = "completed"
	ScanStatusFailed    ScanStatus = "failed"
	ScanStatusCancelled ScanStatus = "cancelled"
)

type ViolationImpact string

const (
	ViolationImpactCritical ViolationImpact = "critical"
	ViolationImpactSerious  ViolationImpact = "serious"
	ViolationImpactModerate ViolationImpact = "moderate"
	ViolationImpactMinor    ViolationImpact = "minor"
)

type ReviewStatus string

const (
	ReviewStatusPending       ReviewStatus = "pending"
	ReviewStatusPassed        ReviewStatus = "passed"
	ReviewStatusFailed        ReviewStatus = "failed"
	ReviewStatusNotApplicable ReviewStatus = "not_applicable"
)

type ReportFormat string

const (
	ReportFormatJSON ReportFormat = "json"
	ReportFormatPDF  ReportFormat = "pdf"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Session struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	TokenHash  string    `json:"-"`
	CSRFHash   string    `json:"-"`
	UserAgent  string    `json:"user_agent"`
	IPAddress  string    `json:"ip_address"`
	ExpiresAt  time.Time `json:"expires_at"`
	LastUsedAt time.Time `json:"last_used_at"`
	CreatedAt  time.Time `json:"created_at"`
}

type Project struct {
	ID          uuid.UUID `json:"id"`
	OwnerID     uuid.UUID `json:"owner_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Scan struct {
	ID             uuid.UUID  `json:"id"`
	ProjectID      uuid.UUID  `json:"project_id"`
	CreatedBy      uuid.UUID  `json:"created_by"`
	URL            string     `json:"url"`
	Status         ScanStatus `json:"status"`
	PageTitle      string     `json:"page_title"`
	AutomatedScore int16      `json:"automated_score"`
	CriticalCount  int32      `json:"critical_count"`
	SeriousCount   int32      `json:"serious_count"`
	ModerateCount  int32      `json:"moderate_count"`
	MinorCount     int32      `json:"minor_count"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	DurationMS     int64      `json:"duration_ms"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type ScannedPage struct {
	ID        uuid.UUID `json:"id"`
	ScanID    uuid.UUID `json:"scan_id"`
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	Language  string    `json:"language"`
	CreatedAt time.Time `json:"created_at"`
}

type Violation struct {
	ID            uuid.UUID       `json:"id"`
	ScannedPageID uuid.UUID       `json:"scanned_page_id"`
	RuleID        string          `json:"rule_id"`
	Impact        ViolationImpact `json:"impact"`
	Description   string          `json:"description"`
	Help          string          `json:"help"`
	HelpURL       string          `json:"help_url"`
	Tags          []string        `json:"tags"`
	ReviewStatus  ReviewStatus    `json:"review_status"`
	Notes         string          `json:"notes"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type ViolationNode struct {
	ID             uuid.UUID `json:"id"`
	ViolationID    uuid.UUID `json:"violation_id"`
	HTML           string    `json:"html"`
	Target         []string  `json:"target"`
	FailureSummary string    `json:"failure_summary"`
	CreatedAt      time.Time `json:"created_at"`
}

type ViolationWithNodes struct {
	Violation Violation       `json:"violation"`
	Nodes     []ViolationNode `json:"nodes"`
}

type ManualReview struct {
	ID        uuid.UUID    `json:"id"`
	ScanID    uuid.UUID    `json:"scan_id"`
	Status    ReviewStatus `json:"status"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type ManualReviewItem struct {
	ID             uuid.UUID    `json:"id"`
	ManualReviewID uuid.UUID    `json:"manual_review_id"`
	Criterion      string       `json:"criterion"`
	Instruction    string       `json:"instruction"`
	Status         ReviewStatus `json:"status"`
	Notes          string       `json:"notes"`
	Position       int32        `json:"position"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

type Report struct {
	ID          uuid.UUID    `json:"id"`
	ScanID      uuid.UUID    `json:"scan_id"`
	CreatedBy   uuid.UUID    `json:"created_by"`
	Format      ReportFormat `json:"format"`
	Filename    string       `json:"filename"`
	ContentType string       `json:"content_type"`
	Content     []byte       `json:"-"`
	CreatedAt   time.Time    `json:"created_at"`
}

type CreateUserParams struct {
	ID           uuid.UUID
	Name         string
	Email        string
	PasswordHash string
}

type CreateSessionParams struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	CSRFHash  string
	UserAgent string
	IPAddress string
	ExpiresAt time.Time
}

type CreateProjectParams struct {
	ID          uuid.UUID
	OwnerID     uuid.UUID
	Name        string
	Description string
}

type AddProjectMemberParams struct {
	ProjectID uuid.UUID
	UserID    uuid.UUID
	Role      string
}

type UpdateProjectParams struct {
	ProjectID   uuid.UUID
	UserID      uuid.UUID
	Name        string
	Description string
}

type ProjectUserParams struct {
	ProjectID uuid.UUID
	UserID    uuid.UUID
}

type CreateScanParams struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	CreatedBy uuid.UUID
	URL       string
}

type ScanUserParams struct {
	ScanID uuid.UUID
	UserID uuid.UUID
}

type ListScansByProjectParams struct {
	ProjectID   uuid.UUID
	UserID      uuid.UUID
	ResultLimit int32
}

type CountRecentScansByUserParams struct {
	UserID    uuid.UUID
	SinceTime time.Time
}

type CompleteScanParams struct {
	ID             uuid.UUID
	PageTitle      string
	AutomatedScore int16
	CriticalCount  int32
	SeriousCount   int32
	ModerateCount  int32
	MinorCount     int32
	DurationMS     int64
}

type CreateScannedPageParams struct {
	ID       uuid.UUID
	ScanID   uuid.UUID
	URL      string
	Title    string
	Language string
}

type CreateViolationParams struct {
	ID            uuid.UUID
	ScannedPageID uuid.UUID
	RuleID        string
	Impact        ViolationImpact
	Description   string
	Help          string
	HelpURL       string
	Tags          []string
}

type CreateViolationNodeParams struct {
	ID             uuid.UUID
	ViolationID    uuid.UUID
	HTML           string
	Target         []string
	FailureSummary string
}

type ViolationUserParams struct {
	ViolationID uuid.UUID
	UserID      uuid.UUID
}

type UpdateViolationReviewParams struct {
	ViolationID  uuid.UUID
	UserID       uuid.UUID
	ReviewStatus ReviewStatus
	Notes        string
}

type CreateManualReviewParams struct {
	ID     uuid.UUID
	ScanID uuid.UUID
}

type CreateManualReviewItemParams struct {
	ID             uuid.UUID
	ManualReviewID uuid.UUID
	Criterion      string
	Instruction    string
	Position       int32
}

type UpdateManualReviewItemParams struct {
	ItemID uuid.UUID
	UserID uuid.UUID
	Status ReviewStatus
	Notes  string
}

type CreateReportParams struct {
	ID          uuid.UUID
	ScanID      uuid.UUID
	CreatedBy   uuid.UUID
	Format      ReportFormat
	Filename    string
	ContentType string
	Content     []byte
}

type ReportUserParams struct {
	ReportID uuid.UUID
	UserID   uuid.UUID
}

type CreateActivityLogParams struct {
	ID         uuid.UUID
	UserID     *uuid.UUID
	ProjectID  *uuid.UUID
	Action     string
	EntityType string
	EntityID   *uuid.UUID
	Metadata   json.RawMessage
}
