package httpapi

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	db "github.com/ki1bot/aksesibilitas-website/internal/database/db"
	taskqueue "github.com/ki1bot/aksesibilitas-website/internal/queue"
	"github.com/ki1bot/aksesibilitas-website/internal/security"
)

var manualReviewTemplates = []struct {
	Criterion   string
	Instruction string
}{
	{
		Criterion:   "Navigasi hanya dengan keyboard",
		Instruction: "Gunakan Tab, Shift+Tab, Enter, Space, dan tombol panah. Pastikan seluruh kontrol dapat digunakan tanpa mouse.",
	},
	{
		Criterion:   "Indikator fokus terlihat",
		Instruction: "Pastikan komponen interaktif menampilkan indikator fokus yang jelas dan tidak tertutup elemen lain.",
	},
	{
		Criterion:   "Urutan fokus masuk akal",
		Instruction: "Periksa bahwa perpindahan fokus mengikuti urutan visual dan urutan membaca halaman.",
	},
	{
		Criterion:   "Pembesaran halaman hingga 200 persen",
		Instruction: "Perbesar halaman hingga 200 persen dan pastikan konten tetap dapat dibaca serta digunakan.",
	},
	{
		Criterion:   "Pengujian pembaca layar",
		Instruction: "Periksa nama, peran, status, heading, landmark, dan pesan kesalahan menggunakan pembaca layar.",
	},
	{
		Criterion:   "Makna tidak bergantung pada warna",
		Instruction: "Pastikan informasi penting tidak disampaikan hanya melalui perbedaan warna.",
	},
}

func (server *Server) listScans(
	writer http.ResponseWriter,
	request *http.Request,
) {
	principal := principalFromContext(
		request.Context(),
	)

	resultLimit := int32(50)

	if rawLimit := request.URL.Query().Get(
		"limit",
	); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err == nil && parsed >= 1 && parsed <= 100 {
			resultLimit = int32(parsed)
		}
	}

	projectValue := request.URL.Query().Get(
		"project_id",
	)

	var scans []db.Scan
	var err error

	if projectValue == "" {
		scans, err = server.queries.ListScansByUser(
			request.Context(),
			principal.User.ID,
			resultLimit,
		)
	} else {
		projectID, parseErr :=
			uuid.Parse(projectValue)

		if parseErr != nil {
			writeError(
				writer,
				http.StatusBadRequest,
				"invalid_project_id",
				"Format project ID tidak valid",
			)
			return
		}

		scans, err =
			server.queries.ListScansByProject(
				request.Context(),
				db.ListScansByProjectParams{
					ProjectID:   projectID,
					UserID:      principal.User.ID,
					ResultLimit: resultLimit,
				},
			)
	}

	if err != nil {
		writeInternalError(writer)
		return
	}

	writeJSON(writer, http.StatusOK, scans)
}

func (server *Server) createScan(
	writer http.ResponseWriter,
	request *http.Request,
) {
	var input scanRequest

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

	projectID, err := uuid.Parse(input.ProjectID)
	if err != nil {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_project_id",
			"Format project ID tidak valid",
		)
		return
	}

	principal := principalFromContext(
		request.Context(),
	)

	_, err = server.queries.GetProjectForUser(
		request.Context(),
		db.ProjectUserParams{
			ProjectID: projectID,
			UserID:    principal.User.ID,
		},
	)
	if err != nil {
		writeDatabaseLookupError(writer, err)
		return
	}

	count, err :=
		server.queries.CountRecentScansByUser(
			request.Context(),
			db.CountRecentScansByUserParams{
				UserID: principal.User.ID,
				SinceTime: time.Now().Add(
					-time.Minute,
				),
			},
		)
	if err != nil {
		writeInternalError(writer)
		return
	}

	if count >= 5 {
		writeError(
			writer,
			http.StatusTooManyRequests,
			"scan_rate_limit",
			"Maksimal lima scan per menit",
		)
		return
	}

	normalizedURL, err :=
		security.ValidatePublicHTTPURL(
			request.Context(),
			input.URL,
		)
	if err != nil {
		writeError(
			writer,
			http.StatusBadRequest,
			"invalid_url",
			err.Error(),
		)
		return
	}

	transaction, err := server.pool.Begin(
		request.Context(),
	)
	if err != nil {
		writeInternalError(writer)
		return
	}
	defer transaction.Rollback(request.Context())

	queries := server.queries.WithTx(transaction)

	scan, err := queries.CreateScan(
		request.Context(),
		db.CreateScanParams{
			ID:        uuid.New(),
			ProjectID: projectID,
			CreatedBy: principal.User.ID,
			URL:       normalizedURL,
		},
	)
	if err != nil {
		writeInternalError(writer)
		return
	}

	manualReview, err :=
		queries.CreateManualReview(
			request.Context(),
			db.CreateManualReviewParams{
				ID:     uuid.New(),
				ScanID: scan.ID,
			},
		)
	if err != nil {
		writeInternalError(writer)
		return
	}

	for index, template := range manualReviewTemplates {
		_, err := queries.CreateManualReviewItem(
			request.Context(),
			db.CreateManualReviewItemParams{
				ID:             uuid.New(),
				ManualReviewID: manualReview.ID,
				Criterion:      template.Criterion,
				Instruction:    template.Instruction,
				Position:       int32(index + 1),
			},
		)
		if err != nil {
			writeInternalError(writer)
			return
		}
	}

	if err := transaction.Commit(
		request.Context(),
	); err != nil {
		writeInternalError(writer)
		return
	}

	task, options, err :=
		taskqueue.NewAccessibilityScanTask(
			scan.ID.String(),
			scan.URL,
			server.cfg.ScanQueue,
			server.cfg.ScanTimeout,
		)
	if err != nil {
		server.failQueuedScan(
			request.Context(),
			scan.ID,
			"Gagal membuat pekerjaan scan",
		)

		writeInternalError(writer)
		return
	}

	if _, err := server.queueClient.Enqueue(
		task,
		options...,
	); err != nil {
		server.failQueuedScan(
			request.Context(),
			scan.ID,
			"Gagal memasukkan scan ke antrean",
		)

		writeInternalError(writer)
		return
	}

	writeJSON(
		writer,
		http.StatusCreated,
		scan,
	)
}

func (server *Server) getScan(
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

	writeJSON(writer, http.StatusOK, scan)
}

func (server *Server) cancelScan(
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

	scan, err := server.queries.CancelScan(
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

	writeJSON(writer, http.StatusOK, scan)
}

func (server *Server) retryScan(
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

	transaction, err := server.pool.Begin(
		request.Context(),
	)
	if err != nil {
		writeInternalError(writer)
		return
	}
	defer transaction.Rollback(request.Context())

	queries := server.queries.WithTx(transaction)

	scan, err := queries.ResetScanForRetry(
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

	if err := queries.DeleteScanResults(
		request.Context(),
		scanID,
	); err != nil {
		writeInternalError(writer)
		return
	}

	if err := queries.DeleteScanReports(
		request.Context(),
		scanID,
	); err != nil {
		writeInternalError(writer)
		return
	}

	if err := queries.ResetManualReview(
		request.Context(),
		scanID,
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

	task, options, err :=
		taskqueue.NewAccessibilityScanTask(
			scan.ID.String(),
			scan.URL,
			server.cfg.ScanQueue,
			server.cfg.ScanTimeout,
		)
	if err != nil {
		server.failQueuedScan(
			request.Context(),
			scan.ID,
			"Gagal membuat pekerjaan retry",
		)

		writeInternalError(writer)
		return
	}

	if _, err := server.queueClient.Enqueue(
		task,
		options...,
	); err != nil {
		server.failQueuedScan(
			request.Context(),
			scan.ID,
			"Gagal memasukkan retry ke antrean",
		)

		writeInternalError(writer)
		return
	}

	writeJSON(writer, http.StatusOK, scan)
}

func (server *Server) deleteScan(
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

	err := server.queries.DeleteScanForUser(
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

	writer.WriteHeader(http.StatusNoContent)
}

func (server *Server) failQueuedScan(
	ctx context.Context,
	scanID uuid.UUID,
	message string,
) {
	_, err := server.queries.FailScan(
		ctx,
		scanID,
		message,
	)
	if err != nil {
		log.Printf(
			"gagal mengubah status scan %s: %v",
			scanID,
			err,
		)
	}
}
