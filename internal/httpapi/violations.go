package httpapi

import (
	"net/http"
	"strings"

	db "github.com/ki1bot/aksesibilitas-website/internal/database/db"
)

func (server *Server) listViolations(
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

	principal := principalFromContext(
		request.Context(),
	)

	violations, err :=
		server.queries.ListViolationsForScan(
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

	writeJSON(
		writer,
		http.StatusOK,
		violations,
	)
}

func (server *Server) getViolation(
	writer http.ResponseWriter,
	request *http.Request,
) {
	violationID, ok := parseUUIDParam(
		writer,
		request,
		"violationId",
	)
	if !ok {
		return
	}

	principal := principalFromContext(
		request.Context(),
	)

	violation, err :=
		server.queries.GetViolationForUser(
			request.Context(),
			db.ViolationUserParams{
				ViolationID: violationID,
				UserID:      principal.User.ID,
			},
		)
	if err != nil {
		writeDatabaseLookupError(writer, err)
		return
	}

	nodes, err :=
		server.queries.ListViolationNodes(
			request.Context(),
			violation.ID,
		)
	if err != nil {
		writeInternalError(writer)
		return
	}

	writeJSON(
		writer,
		http.StatusOK,
		violationResponse{
			Violation: violation,
			Nodes:     nodes,
		},
	)
}

func (server *Server) updateViolation(
	writer http.ResponseWriter,
	request *http.Request,
) {
	violationID, ok := parseUUIDParam(
		writer,
		request,
		"violationId",
	)
	if !ok {
		return
	}

	var input reviewRequest

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

	status, valid := parseReviewStatus(
		input.Status,
	)
	if !valid || len(input.Notes) > 5000 {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_review",
			"Status atau catatan review tidak valid",
		)
		return
	}

	principal := principalFromContext(
		request.Context(),
	)

	violation, err :=
		server.queries.UpdateViolationReview(
			request.Context(),
			db.UpdateViolationReviewParams{
				ViolationID:  violationID,
				UserID:       principal.User.ID,
				ReviewStatus: status,
				Notes:        strings.TrimSpace(input.Notes),
			},
		)
	if err != nil {
		writeDatabaseLookupError(writer, err)
		return
	}

	writeJSON(writer, http.StatusOK, violation)
}
