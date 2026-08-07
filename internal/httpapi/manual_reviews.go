package httpapi

import (
	"net/http"
	"strings"

	db "github.com/ki1bot/aksesibilitas-website/internal/database/db"
)

func (server *Server) getManualReview(
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

	review, err :=
		server.queries.GetManualReviewForScan(
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

	items, err :=
		server.queries.ListManualReviewItems(
			request.Context(),
			review.ID,
		)
	if err != nil {
		writeInternalError(writer)
		return
	}

	writeJSON(
		writer,
		http.StatusOK,
		manualReviewResponse{
			Review: review,
			Items:  items,
		},
	)
}

func (server *Server) updateManualReviewItem(
	writer http.ResponseWriter,
	request *http.Request,
) {
	itemID, ok := parseUUIDParam(
		writer,
		request,
		"itemId",
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

	transaction, err := server.pool.Begin(
		request.Context(),
	)
	if err != nil {
		writeInternalError(writer)
		return
	}
	defer transaction.Rollback(request.Context())

	queries := server.queries.WithTx(transaction)

	item, err := queries.UpdateManualReviewItem(
		request.Context(),
		db.UpdateManualReviewItemParams{
			ItemID: itemID,
			UserID: principal.User.ID,
			Status: status,
			Notes:  strings.TrimSpace(input.Notes),
		},
	)
	if err != nil {
		writeDatabaseLookupError(writer, err)
		return
	}

	if err := queries.RefreshManualReviewStatus(
		request.Context(),
		item.ManualReviewID,
	); err != nil {
		writeInternalError(writer)
		return
	}

	if err := transaction.Commit(
		request.Context(),
	); err != nil {
		writeInternalError(writer)
		return
	}

	writeJSON(writer, http.StatusOK, item)
}
