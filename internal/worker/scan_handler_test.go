package worker

import (
	"errors"
	"strings"
	"testing"

	db "github.com/ki1bot/aksesibilitas-website/internal/database/db"
)

func TestNormalizeImpact(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected db.ViolationImpact
	}{
		{
			name:     "critical",
			input:    "critical",
			expected: db.ViolationImpactCritical,
		},
		{
			name:     "critical uppercase",
			input:    "CRITICAL",
			expected: db.ViolationImpactCritical,
		},
		{
			name:     "serious",
			input:    "serious",
			expected: db.ViolationImpactSerious,
		},
		{
			name:     "moderate",
			input:    "moderate",
			expected: db.ViolationImpactModerate,
		},
		{
			name:     "minor",
			input:    "minor",
			expected: db.ViolationImpactMinor,
		},
		{
			name:     "empty impact",
			input:    "",
			expected: db.ViolationImpactMinor,
		},
		{
			name:     "unknown impact",
			input:    "unknown",
			expected: db.ViolationImpactMinor,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(
			testCase.name,
			func(t *testing.T) {
				t.Parallel()

				actual := normalizeImpact(testCase.input)

				if actual != testCase.expected {
					t.Fatalf(
						"normalizeImpact(%q) = %q, diharapkan %q",
						testCase.input,
						actual,
						testCase.expected,
					)
				}
			},
		)
	}
}

func TestCalculateScore(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		critical int32
		serious  int32
		moderate int32
		minor    int32
		expected int16
	}{
		{
			name:     "no violations",
			critical: 0,
			serious:  0,
			moderate: 0,
			minor:    0,
			expected: 100,
		},
		{
			name:     "one critical",
			critical: 1,
			serious:  0,
			moderate: 0,
			minor:    0,
			expected: 85,
		},
		{
			name:     "one serious",
			critical: 0,
			serious:  1,
			moderate: 0,
			minor:    0,
			expected: 92,
		},
		{
			name:     "one moderate",
			critical: 0,
			serious:  0,
			moderate: 1,
			minor:    0,
			expected: 96,
		},
		{
			name:     "one minor",
			critical: 0,
			serious:  0,
			moderate: 0,
			minor:    1,
			expected: 99,
		},
		{
			name:     "mixed violations",
			critical: 1,
			serious:  2,
			moderate: 3,
			minor:    4,
			expected: 53,
		},
		{
			name:     "score cannot be negative",
			critical: 10,
			serious:  10,
			moderate: 10,
			minor:    10,
			expected: 0,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(
			testCase.name,
			func(t *testing.T) {
				t.Parallel()

				actual := calculateScore(
					testCase.critical,
					testCase.serious,
					testCase.moderate,
					testCase.minor,
				)

				if actual != testCase.expected {
					t.Fatalf(
						"calculateScore(%d, %d, %d, %d) = %d, diharapkan %d",
						testCase.critical,
						testCase.serious,
						testCase.moderate,
						testCase.minor,
						actual,
						testCase.expected,
					)
				}
			},
		)
	}
}

func TestSanitizeErrorTrimsWhitespace(
	t *testing.T,
) {
	t.Parallel()

	actual := sanitizeError(
		errors.New("   pemindaian gagal   "),
	)

	if actual != "pemindaian gagal" {
		t.Fatalf(
			"sanitizeError menghasilkan %q",
			actual,
		)
	}
}

func TestSanitizeErrorLimitsLength(
	t *testing.T,
) {
	t.Parallel()

	actual := sanitizeError(
		errors.New(strings.Repeat("a", 1200)),
	)

	if len(actual) != 1000 {
		t.Fatalf(
			"panjang hasil = %d, diharapkan 1000",
			len(actual),
		)
	}

	if actual != strings.Repeat("a", 1000) {
		t.Fatal("isi error terpotong tidak sesuai")
	}
}
