package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"

	db "github.com/ki1bot/aksescheck-id/internal/database/db"
	taskqueue "github.com/ki1bot/aksescheck-id/internal/queue"
)

type ScanHandler struct {
	queries db.Querier
}

func NewScanHandler(queries db.Querier) *ScanHandler {
	return &ScanHandler{
		queries: queries,
	}
}

func (handler *ScanHandler) ProcessTask(
	ctx context.Context,
	task *asynq.Task,
) error {
	var payload taskqueue.AccessibilityScanPayload

	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf(
			"payload task tidak valid: %v: %w",
			err,
			asynq.SkipRetry,
		)
	}

	scanID, err := uuid.Parse(payload.ScanID)
	if err != nil {
		return fmt.Errorf(
			"scan ID tidak valid: %v: %w",
			err,
			asynq.SkipRetry,
		)
	}

	_, err = handler.queries.UpdateScanStatus(
		ctx,
		db.UpdateScanStatusParams{
			ID:           scanID,
			Status:       db.ScanStatus("running"),
			ErrorMessage: "",
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf(
				"scan tidak ditemukan: %w",
				asynq.SkipRetry,
			)
		}

		return fmt.Errorf(
			"gagal mengubah status menjadi running: %w",
			err,
		)
	}

	message := "Mesin chromedp dan axe-core belum diterapkan pada tahap 3-5"

	_, err = handler.queries.UpdateScanStatus(
		ctx,
		db.UpdateScanStatusParams{
			ID:           scanID,
			Status:       db.ScanStatus("failed"),
			ErrorMessage: message,
		},
	)
	if err != nil {
		return fmt.Errorf(
			"gagal mengubah status menjadi failed: %w",
			err,
		)
	}

	log.Printf(
		"scan %s untuk %s diterima worker",
		payload.ScanID,
		payload.URL,
	)

	return nil
}
