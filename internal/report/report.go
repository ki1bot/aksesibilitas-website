package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ki1bot/aksesibilitas-website/internal/database/db"
)

type Payload struct {
	GeneratedAt  time.Time               `json:"generated_at"`
	Disclaimer   string                  `json:"disclaimer"`
	Scan         db.Scan                 `json:"scan"`
	Violations   []db.ViolationWithNodes `json:"violations"`
	ManualReview []db.ManualReviewItem   `json:"manual_review"`
}

func BuildJSON(
	scan db.Scan,
	violations []db.ViolationWithNodes,
	manualItems []db.ManualReviewItem,
) ([]byte, error) {
	return json.MarshalIndent(
		Payload{
			GeneratedAt:  time.Now(),
			Disclaimer:   "Hasil otomatis hanya mencakup aturan yang dapat diperiksa mesin. Pemeriksaan manual tetap diperlukan untuk menentukan kesesuaian WCAG.",
			Scan:         scan,
			Violations:   violations,
			ManualReview: manualItems,
		},
		"",
		"  ",
	)
}

func BuildPDF(
	scan db.Scan,
	violations []db.ViolationWithNodes,
	manualItems []db.ManualReviewItem,
) ([]byte, error) {
	lines := make([]string, 0)

	lines = append(
		lines,
		"AksesCheck ID",
		"Laporan Pemeriksaan Aksesibilitas Website",
		"",
		"URL: "+scan.URL,
		"Judul halaman: "+emptyFallback(scan.PageTitle, "-"),
		"Status: "+string(scan.Status),
		"Skor otomatis: "+strconv.Itoa(int(scan.AutomatedScore))+"/100",
		"Durasi: "+strconv.FormatInt(scan.DurationMS, 10)+" ms",
		"Waktu scan: "+scan.CreatedAt.Format("02 January 2006 15:04 MST"),
		"",
		"Ringkasan pelanggaran",
		"Critical: "+strconv.Itoa(int(scan.CriticalCount)),
		"Serious: "+strconv.Itoa(int(scan.SeriousCount)),
		"Moderate: "+strconv.Itoa(int(scan.ModerateCount)),
		"Minor: "+strconv.Itoa(int(scan.MinorCount)),
		"",
		"Catatan penting",
		"Hasil otomatis hanya mencakup aturan yang dapat diperiksa mesin.",
		"Pemeriksaan manual tetap diperlukan untuk menentukan kesesuaian WCAG.",
	)

	sort.SliceStable(
		violations,
		func(left, right int) bool {
			return impactWeight(
				violations[left].Violation.Impact,
			) < impactWeight(
				violations[right].Violation.Impact,
			)
		},
	)

	for index, item := range violations {
		violation := item.Violation

		lines = append(
			lines,
			"",
			fmt.Sprintf(
				"Pelanggaran %d: %s",
				index+1,
				violation.RuleID,
			),
			"Dampak: "+string(violation.Impact),
			"Masalah: "+violation.Description,
			"Saran: "+violation.Help,
		)

		if violation.HelpURL != "" {
			lines = append(
				lines,
				"Referensi: "+violation.HelpURL,
			)
		}

		for nodeIndex, node := range item.Nodes {
			lines = append(
				lines,
				fmt.Sprintf(
					"Node %d: %s",
					nodeIndex+1,
					strings.Join(node.Target, " "),
				),
				"HTML: "+node.HTML,
				"Detail: "+node.FailureSummary,
			)
		}
	}

	lines = append(
		lines,
		"",
		"Pemeriksaan manual",
	)

	for _, item := range manualItems {
		lines = append(
			lines,
			fmt.Sprintf(
				"%d. %s [%s]",
				item.Position,
				item.Criterion,
				item.Status,
			),
			item.Instruction,
		)

		if item.Notes != "" {
			lines = append(
				lines,
				"Catatan: "+item.Notes,
			)
		}
	}

	return renderPDF(lines), nil
}

func renderPDF(lines []string) []byte {
	wrapped := make([]string, 0)

	for _, line := range lines {
		wrapped = append(
			wrapped,
			wrapLine(line, 86)...,
		)
	}

	const linesPerPage = 48

	pageCount := (len(wrapped) + linesPerPage - 1) /
		linesPerPage

	if pageCount == 0 {
		pageCount = 1
	}

	objects := make([][]byte, 3+pageCount*2)

	kids := make([]string, 0, pageCount)

	for pageIndex := 0; pageIndex < pageCount; pageIndex++ {
		pageObject := 4 + pageIndex*2
		contentObject := pageObject + 1

		kids = append(
			kids,
			fmt.Sprintf("%d 0 R", pageObject),
		)

		start := pageIndex * linesPerPage
		end := start + linesPerPage

		if end > len(wrapped) {
			end = len(wrapped)
		}

		content := buildPageContent(
			wrapped[start:end],
			pageIndex+1,
			pageCount,
		)

		objects[pageObject-1] = []byte(
			fmt.Sprintf(
				"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 3 0 R >> >> /Contents %d 0 R >>",
				contentObject,
			),
		)

		objects[contentObject-1] = []byte(
			fmt.Sprintf(
				"<< /Length %d >>\nstream\n%s\nendstream",
				len(content),
				content,
			),
		)
	}

	objects[0] = []byte(
		"<< /Type /Catalog /Pages 2 0 R >>",
	)

	objects[1] = []byte(
		fmt.Sprintf(
			"<< /Type /Pages /Count %d /Kids [%s] >>",
			pageCount,
			strings.Join(kids, " "),
		),
	)

	objects[2] = []byte(
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
	)

	var output bytes.Buffer

	output.WriteString("%PDF-1.4\n")
	output.Write([]byte{0x25, 0xE2, 0xE3, 0xCF, 0xD3, '\n'})

	offsets := make([]int, len(objects)+1)

	for index, object := range objects {
		offsets[index+1] = output.Len()

		output.WriteString(
			fmt.Sprintf("%d 0 obj\n", index+1),
		)

		output.Write(object)
		output.WriteString("\nendobj\n")
	}

	xrefOffset := output.Len()

	output.WriteString(
		fmt.Sprintf(
			"xref\n0 %d\n",
			len(objects)+1,
		),
	)

	output.WriteString("0000000000 65535 f \n")

	for index := 1; index < len(offsets); index++ {
		output.WriteString(
			fmt.Sprintf(
				"%010d 00000 n \n",
				offsets[index],
			),
		)
	}

	output.WriteString(
		fmt.Sprintf(
			"trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF",
			len(objects)+1,
			xrefOffset,
		),
	)

	return output.Bytes()
}

func buildPageContent(
	lines []string,
	pageNumber int,
	pageCount int,
) string {
	var content strings.Builder

	content.WriteString("BT\n/F1 10 Tf\n48 792 Td\n")

	for index, line := range lines {
		if index > 0 {
			content.WriteString("0 -15 Td\n")
		}

		content.WriteString("(")
		content.WriteString(escapePDFText(line))
		content.WriteString(") Tj\n")
	}

	content.WriteString("ET\n")
	content.WriteString("BT\n/F1 8 Tf\n")
	content.WriteString("260 24 Td\n")
	content.WriteString(
		fmt.Sprintf(
			"(Halaman %d dari %d) Tj\n",
			pageNumber,
			pageCount,
		),
	)
	content.WriteString("ET")

	return content.String()
}

func wrapLine(value string, limit int) []string {
	if value == "" {
		return []string{""}
	}

	words := strings.Fields(value)
	if len(words) == 0 {
		return []string{""}
	}

	lines := make([]string, 0)
	current := words[0]

	for _, word := range words[1:] {
		if len(current)+1+len(word) <= limit {
			current += " " + word
			continue
		}

		lines = append(lines, current)
		current = word
	}

	lines = append(lines, current)

	return lines
}

func escapePDFText(value string) string {
	var builder strings.Builder

	for _, character := range value {
		switch character {
		case '\\', '(', ')':
			builder.WriteByte('\\')
			builder.WriteRune(character)
		default:
			if character >= 32 && character <= 255 {
				builder.WriteRune(character)
			} else {
				builder.WriteByte('?')
			}
		}
	}

	return builder.String()
}

func emptyFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}

	return value
}

func impactWeight(impact db.ViolationImpact) int {
	switch impact {
	case db.ViolationImpactCritical:
		return 1
	case db.ViolationImpactSerious:
		return 2
	case db.ViolationImpactModerate:
		return 3
	default:
		return 4
	}
}
