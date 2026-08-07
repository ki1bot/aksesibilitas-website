package httpapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	db "github.com/ki1bot/aksesibilitas-website/internal/database/db"
	reportbuilder "github.com/ki1bot/aksesibilitas-website/internal/report"
)

func (server *Server) createReport(
	writer http.ResponseWriter,
	request *http.Request,
) {
	scanID, ok := parseUUIDParam(
		writer,
		request,
		"scanId",
	)
	if !ok {
		return
	}

	var input reportRequest

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

	format := db.ReportFormat(
		strings.ToLower(
			strings.TrimSpace(input.Format),
		),
	)

	if format != db.ReportFormatJSON &&
		format != db.ReportFormatPDF {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_report_format",
			"Format laporan harus json atau pdf",
		)
		return
	}

	principal := principalFromContext(
		request.Context(),
	)

	scan, err := server.queries.GetScanForUser(
		request.Context(),
		db.ScanUserParams{
			ScanID: scanID,
			UserID: principal.User.ID,
		},
	)
	if err != nil {
		writeDatabaseLookupError(writer, err)
		return
	}

	if scan.Status != db.ScanStatusCompleted {
		writeError(
			writer,
			http.StatusConflict,
			"scan_not_completed",
			"Laporan hanya dapat dibuat setelah scan selesai",
		)
		return
	}

	violations, err :=
		server.loadViolationsWithNodes(
			request.Context(),
			scanID,
			principal.User.ID,
		)
	if err != nil {
		writeInternalError(writer)
		return
	}

	review, err :=
		server.queries.GetManualReviewForScan(
			request.Context(),
			db.ScanUserParams{
				ScanID: scanID,
				UserID: principal.User.ID,
			},
		)
	if err != nil {
		writeInternalError(writer)
		return
	}

	manualItems, err :=
		server.queries.ListManualReviewItems(
			request.Context(),
			review.ID,
		)
	if err != nil {
		writeInternalError(writer)
		return
	}

	var content []byte
	var filename string
	var contentType string

	switch format {
	case db.ReportFormatPDF:
		content, err = reportbuilder.BuildPDF(
			scan,
			violations,
			manualItems,
		)

		filename = "aksescheck-" +
			scan.ID.String() +
			".pdf"

		contentType = "application/pdf"
	default:
		content, err = reportbuilder.BuildJSON(
			scan,
			violations,
			manualItems,
		)

		filename = "aksescheck-" +
			scan.ID.String() +
			".json"

		contentType = "application/json"
	}

	if err != nil {
		writeInternalError(writer)
		return
	}

	report, err := server.queries.CreateReport(
		request.Context(),
		db.CreateReportParams{
			ID:          uuid.New(),
			ScanID:      scan.ID,
			CreatedBy:   principal.User.ID,
			Format:      format,
			Filename:    filename,
			ContentType: contentType,
			Content:     content,
		},
	)
	if err != nil {
		writeInternalError(writer)
		return
	}

	writeJSON(
		writer,
		http.StatusCreated,
		toReportResponse(report),
	)
}

func (server *Server) getReport(
	writer http.ResponseWriter,
	request *http.Request,
) {
	report, ok := server.findReport(
		writer,
		request,
	)
	if !ok {
		return
	}

	writeJSON(
		writer,
		http.StatusOK,
		toReportResponse(report),
	)
}

func (server *Server) downloadReport(
	writer http.ResponseWriter,
	request *http.Request,
) {
	report, ok := server.findReport(
		writer,
		request,
	)
	if !ok {
		return
	}

	writer.Header().Set(
		"Content-Type",
		report.ContentType,
	)

	writer.Header().Set(
		"Content-Disposition",
		`attachment; filename="`+
			strings.ReplaceAll(
				report.Filename,
				`"`,
				"",
			)+
			`"`,
	)

	writer.Header().Set(
		"Content-Length",
		strconv.Itoa(len(report.Content)),
	)

	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(report.Content)
}

func (server *Server) findReport(
	writer http.ResponseWriter,
	request *http.Request,
) (db.Report, bool) {
	reportID, ok := parseUUIDParam(
		writer,
		request,
		"reportId",
	)
	if !ok {
		return db.Report{}, false
	}

	principal := principalFromContext(
		request.Context(),
	)

	report, err :=
		server.queries.GetReportForUser(
			request.Context(),
			db.ReportUserParams{
				ReportID: reportID,
				UserID:   principal.User.ID,
			},
		)
	if err != nil {
		writeDatabaseLookupError(writer, err)
		return db.Report{}, false
	}

	return report, true
}

func (server *Server) loadViolationsWithNodes(
	ctx context.Context,
	scanID uuid.UUID,
	userID uuid.UUID,
) ([]db.ViolationWithNodes, error) {
	violations, err :=
		server.queries.ListViolationsForScan(
			ctx,
			db.ScanUserParams{
				ScanID: scanID,
				UserID: userID,
			},
		)
	if err != nil {
		return nil, err
	}

	result := make(
		[]db.ViolationWithNodes,
		0,
		len(violations),
	)

	for _, violation := range violations {
		nodes, nodeErr :=
			server.queries.ListViolationNodes(
				ctx,
				violation.ID,
			)
		if nodeErr != nil {
			return nil, nodeErr
		}

		result = append(
			result,
			db.ViolationWithNodes{
				Violation: violation,
				Nodes:     nodes,
			},
		)
	}

	return result, nil
}

func toReportResponse(
	report db.Report,
) reportResponse {
	return reportResponse{
		ID:          report.ID,
		ScanID:      report.ScanID,
		Format:      report.Format,
		Filename:    report.Filename,
		ContentType: report.ContentType,
		CreatedAt:   report.CreatedAt,
	}
}
