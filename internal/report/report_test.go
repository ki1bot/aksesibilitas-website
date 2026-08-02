package report

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	db "github.com/ki1bot/aksesibilitas-website/internal/database/db"
)

func TestBuildJSON(t *testing.T) {
	t.Parallel()

	scan, violations, manualItems :=
		createReportTestData()

	content, err := BuildJSON(
		scan,
		violations,
		manualItems,
	)
	if err != nil {
		t.Fatalf(
			"BuildJSON menghasilkan error: %v",
			err,
		)
	}

	var payload Payload

	if err := json.Unmarshal(
		content,
		&payload,
	); err != nil {
		t.Fatalf(
			"hasil BuildJSON bukan JSON valid: %v",
			err,
		)
	}

	if payload.GeneratedAt.IsZero() {
		t.Fatal("generated_at tidak boleh kosong")
	}

	if payload.Scan.ID != scan.ID {
		t.Fatalf(
			"scan ID = %s, diharapkan %s",
			payload.Scan.ID,
			scan.ID,
		)
	}

	if payload.Scan.URL != scan.URL {
		t.Fatalf(
			"scan URL = %q, diharapkan %q",
			payload.Scan.URL,
			scan.URL,
		)
	}

	if len(payload.Violations) != 2 {
		t.Fatalf(
			"jumlah violations = %d, diharapkan 2",
			len(payload.Violations),
		)
	}

	if len(payload.ManualReview) != 1 {
		t.Fatalf(
			"jumlah manual review = %d, diharapkan 1",
			len(payload.ManualReview),
		)
	}

	if !strings.Contains(
		payload.Disclaimer,
		"Pemeriksaan manual tetap diperlukan",
	) {
		t.Fatalf(
			"disclaimer tidak sesuai: %q",
			payload.Disclaimer,
		)
	}
}

func TestBuildPDF(t *testing.T) {
	t.Parallel()

	scan, violations, manualItems :=
		createReportTestData()

	content, err := BuildPDF(
		scan,
		violations,
		manualItems,
	)
	if err != nil {
		t.Fatalf(
			"BuildPDF menghasilkan error: %v",
			err,
		)
	}

	if !bytes.HasPrefix(
		content,
		[]byte("%PDF-1.4"),
	) {
		t.Fatal("hasil BuildPDF tidak memiliki header PDF")
	}

	if !bytes.HasSuffix(
		content,
		[]byte("%%EOF"),
	) {
		t.Fatal("hasil BuildPDF tidak memiliki penutup PDF")
	}

	pdfText := string(content)

	expectedTexts := []string{
		"AksesCheck ID",
		"Laporan Pemeriksaan Aksesibilitas Website",
		"https://example.com",
		"Skor otomatis: 76/100",
		"color-contrast",
		"button-name",
		"Pemeriksaan manual",
		"Navigasi keyboard",
	}

	for _, expectedText := range expectedTexts {
		if !strings.Contains(
			pdfText,
			expectedText,
		) {
			t.Fatalf(
				"PDF tidak mengandung %q",
				expectedText,
			)
		}
	}
}

func TestRenderPDFWithEmptyLines(
	t *testing.T,
) {
	t.Parallel()

	content := renderPDF(nil)

	if !bytes.HasPrefix(
		content,
		[]byte("%PDF-1.4"),
	) {
		t.Fatal("renderPDF tidak menghasilkan header PDF")
	}

	if !bytes.Contains(
		content,
		[]byte("/Count 1"),
	) {
		t.Fatal("PDF kosong harus tetap memiliki satu halaman")
	}
}

func TestWrapLine(t *testing.T) {
	t.Parallel()

	actual := wrapLine(
		"satu dua tiga empat",
		8,
	)

	expected := []string{
		"satu dua",
		"tiga",
		"empat",
	}

	if !reflect.DeepEqual(
		actual,
		expected,
	) {
		t.Fatalf(
			"wrapLine menghasilkan %#v, diharapkan %#v",
			actual,
			expected,
		)
	}
}

func TestWrapLineWithEmptyValue(
	t *testing.T,
) {
	t.Parallel()

	actual := wrapLine("", 10)
	expected := []string{""}

	if !reflect.DeepEqual(
		actual,
		expected,
	) {
		t.Fatalf(
			"wrapLine menghasilkan %#v, diharapkan %#v",
			actual,
			expected,
		)
	}
}

func TestEscapePDFText(t *testing.T) {
	t.Parallel()

	actual := escapePDFText(
		"teks (uji) \\ data",
	)

	expected := "teks \\(uji\\) \\\\ data"

	if actual != expected {
		t.Fatalf(
			"escapePDFText menghasilkan %q, diharapkan %q",
			actual,
			expected,
		)
	}
}

func TestEmptyFallback(t *testing.T) {
	t.Parallel()

	if actual := emptyFallback(
		"   ",
		"-",
	); actual != "-" {
		t.Fatalf(
			"emptyFallback menghasilkan %q",
			actual,
		)
	}

	if actual := emptyFallback(
		"Judul halaman",
		"-",
	); actual != "Judul halaman" {
		t.Fatalf(
			"emptyFallback menghasilkan %q",
			actual,
		)
	}
}

func TestImpactWeight(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		impact   db.ViolationImpact
		expected int
	}{
		{
			impact:   db.ViolationImpactCritical,
			expected: 1,
		},
		{
			impact:   db.ViolationImpactSerious,
			expected: 2,
		},
		{
			impact:   db.ViolationImpactModerate,
			expected: 3,
		},
		{
			impact:   db.ViolationImpactMinor,
			expected: 4,
		},
	}

	for _, testCase := range testCases {
		actual := impactWeight(testCase.impact)

		if actual != testCase.expected {
			t.Fatalf(
				"impactWeight(%q) = %d, diharapkan %d",
				testCase.impact,
				actual,
				testCase.expected,
			)
		}
	}
}

func createReportTestData() (
	db.Scan,
	[]db.ViolationWithNodes,
	[]db.ManualReviewItem,
) {
	scanID := uuid.New()
	userID := uuid.New()
	projectID := uuid.New()
	pageID := uuid.New()
	criticalViolationID := uuid.New()
	minorViolationID := uuid.New()
	manualReviewID := uuid.New()

	createdAt := time.Date(
		2026,
		time.August,
		3,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	scan := db.Scan{
		ID:             scanID,
		ProjectID:      projectID,
		CreatedBy:      userID,
		URL:            "https://example.com",
		Status:         db.ScanStatusCompleted,
		PageTitle:      "Example Domain",
		AutomatedScore: 76,
		CriticalCount:  1,
		SeriousCount:   0,
		ModerateCount:  1,
		MinorCount:     5,
		DurationMS:     1250,
		CreatedAt:      createdAt,
		UpdatedAt:      createdAt,
	}

	violations := []db.ViolationWithNodes{
		{
			Violation: db.Violation{
				ID:            minorViolationID,
				ScannedPageID: pageID,
				RuleID:        "button-name",
				Impact:        db.ViolationImpactMinor,
				Description:   "Button harus memiliki nama yang dapat diakses.",
				Help:          "Tambahkan nama yang dapat diakses ke button.",
				HelpURL:       "https://dequeuniversity.com/rules/axe/button-name",
				Tags:          []string{"wcag2a", "wcag412"},
				ReviewStatus:  db.ReviewStatusPending,
				Notes:         "",
				CreatedAt:     createdAt,
				UpdatedAt:     createdAt,
			},
			Nodes: []db.ViolationNode{
				{
					ID:             uuid.New(),
					ViolationID:    minorViolationID,
					HTML:           "<button></button>",
					Target:         []string{"button"},
					FailureSummary: "Perbaiki button agar memiliki nama.",
					CreatedAt:      createdAt,
				},
			},
		},
		{
			Violation: db.Violation{
				ID:            criticalViolationID,
				ScannedPageID: pageID,
				RuleID:        "color-contrast",
				Impact:        db.ViolationImpactCritical,
				Description:   "Kontras warna teks tidak mencukupi.",
				Help:          "Perbaiki rasio kontras warna.",
				HelpURL:       "https://dequeuniversity.com/rules/axe/color-contrast",
				Tags:          []string{"wcag2aa", "wcag143"},
				ReviewStatus:  db.ReviewStatusFailed,
				Notes:         "Kontras harus diperbaiki.",
				CreatedAt:     createdAt,
				UpdatedAt:     createdAt,
			},
			Nodes: []db.ViolationNode{
				{
					ID:             uuid.New(),
					ViolationID:    criticalViolationID,
					HTML:           "<p class=\"muted\">Teks</p>",
					Target:         []string{".muted"},
					FailureSummary: "Rasio kontras berada di bawah batas minimum.",
					CreatedAt:      createdAt,
				},
			},
		},
	}

	manualItems := []db.ManualReviewItem{
		{
			ID:             uuid.New(),
			ManualReviewID: manualReviewID,
			Criterion:      "Navigasi keyboard",
			Instruction:    "Pastikan seluruh kontrol dapat digunakan dengan keyboard.",
			Status:         db.ReviewStatusPassed,
			Notes:          "Seluruh kontrol dapat digunakan dengan Tab dan Enter.",
			Position:       1,
			CreatedAt:      createdAt,
			UpdatedAt:      createdAt,
		},
	}

	return scan, violations, manualItems
}
