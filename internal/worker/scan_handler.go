package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/ki1bot/aksesibilitas-website/internal/database/db"
	"github.com/ki1bot/aksesibilitas-website/internal/scanner"
)

type ScanHandler struct {
	pool    *pgxpool.Pool
	queries *db.Queries
	scanner *scanner.Scanner
}

func NewScanHandler(
	pool *pgxpool.Pool,
	chromePath string,
) *ScanHandler {
	return &ScanHandler{
		pool:    pool,
		queries: db.New(pool),
		scanner: scanner.New(chromePath),
	}
}

func (handler *ScanHandler) Process(
	ctx context.Context,
	scanID uuid.UUID,
) error {
	currentScan, err := handler.queries.GetScanByID(
		ctx,
		scanID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("scan tidak ditemukan")
		}

		return fmt.Errorf(
			"gagal mengambil scan: %w",
			err,
		)
	}

	if currentScan.Status == db.ScanStatusCancelled ||
		currentScan.Status == db.ScanStatusCompleted {
		return nil
	}

	runningScan, err := handler.queries.MarkScanRunning(
		ctx,
		scanID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}

		return fmt.Errorf(
			"gagal mengubah status scan menjadi running: %w",
			err,
		)
	}

	scanContext, cancelScan := context.WithCancel(ctx)
	defer cancelScan()

	watchDone := make(chan struct{})

	go handler.watchCancellation(
		scanContext,
		watchDone,
		scanID,
		cancelScan,
	)

	result, scanErr := handler.scanner.Scan(
		scanContext,
		runningScan.URL,
	)

	close(watchDone)

	if scanErr != nil {
		latestScan, latestErr := handler.queries.GetScanByID(
			context.Background(),
			scanID,
		)

		if latestErr == nil &&
			latestScan.Status == db.ScanStatusCancelled {
			return nil
		}

		message := sanitizeError(scanErr)

		_, failErr := handler.queries.FailScan(
			context.Background(),
			scanID,
			message,
		)

		if failErr != nil &&
			!errors.Is(failErr, pgx.ErrNoRows) {
			log.Printf(
				"gagal menyimpan kegagalan scan %s: %v",
				scanID,
				failErr,
			)
		}

		return fmt.Errorf(
			"pemindaian gagal: %w",
			scanErr,
		)
	}

	transaction, err := handler.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf(
			"gagal memulai transaksi hasil scan: %w",
			err,
		)
	}
	defer transaction.Rollback(ctx)

	queries := handler.queries.WithTx(transaction)

	latestScan, err := queries.GetScanByID(
		ctx,
		scanID,
	)
	if err != nil {
		return err
	}

	if latestScan.Status == db.ScanStatusCancelled {
		return transaction.Commit(ctx)
	}

	if err := queries.DeleteScanResults(
		ctx,
		scanID,
	); err != nil {
		return err
	}

	if err := queries.DeleteScanReports(
		ctx,
		scanID,
	); err != nil {
		return err
	}

	page, err := queries.CreateScannedPage(
		ctx,
		db.CreateScannedPageParams{
			ID:       uuid.New(),
			ScanID:   scanID,
			URL:      result.URL,
			Title:    result.Title,
			Language: result.Language,
		},
	)
	if err != nil {
		return fmt.Errorf(
			"gagal menyimpan halaman hasil scan: %w",
			err,
		)
	}

	var criticalCount int32
	var seriousCount int32
	var moderateCount int32
	var minorCount int32

	for _, resultViolation := range result.Violations {
		impact := normalizeImpact(
			resultViolation.Impact,
		)

		switch impact {
		case db.ViolationImpactCritical:
			criticalCount++
		case db.ViolationImpactSerious:
			seriousCount++
		case db.ViolationImpactModerate:
			moderateCount++
		default:
			minorCount++
		}

		violation, createErr := queries.CreateViolation(
			ctx,
			db.CreateViolationParams{
				ID:            uuid.New(),
				ScannedPageID: page.ID,
				RuleID:        resultViolation.ID,
				Impact:        impact,
				Description:   resultViolation.Description,
				Help:          resultViolation.Help,
				HelpURL:       resultViolation.HelpURL,
				Tags:          resultViolation.Tags,
			},
		)
		if createErr != nil {
			return fmt.Errorf(
				"gagal menyimpan violation %s: %w",
				resultViolation.ID,
				createErr,
			)
		}

		for _, resultNode := range resultViolation.Nodes {
			_, nodeErr := queries.CreateViolationNode(
				ctx,
				db.CreateViolationNodeParams{
					ID:             uuid.New(),
					ViolationID:    violation.ID,
					HTML:           resultNode.HTML,
					Target:         resultNode.Target,
					FailureSummary: resultNode.FailureSummary,
				},
			)
			if nodeErr != nil {
				return fmt.Errorf(
					"gagal menyimpan node violation: %w",
					nodeErr,
				)
			}
		}
	}

	score := calculateScore(
		criticalCount,
		seriousCount,
		moderateCount,
		minorCount,
	)

	_, err = queries.CompleteScan(
		ctx,
		db.CompleteScanParams{
			ID:             scanID,
			PageTitle:      result.Title,
			AutomatedScore: score,
			CriticalCount:  criticalCount,
			SeriousCount:   seriousCount,
			ModerateCount:  moderateCount,
			MinorCount:     minorCount,
			DurationMS:     result.DurationMS,
		},
	)
	if err != nil {
		return fmt.Errorf(
			"gagal menyelesaikan scan: %w",
			err,
		)
	}

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf(
			"gagal melakukan commit hasil scan: %w",
			err,
		)
	}

	log.Printf(
		"scan %s selesai untuk %s dengan %d violation",
		scanID,
		result.URL,
		len(result.Violations),
	)

	return nil
}

func (handler *ScanHandler) watchCancellation(
	ctx context.Context,
	done <-chan struct{},
	scanID uuid.UUID,
	cancel context.CancelFunc,
) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			currentScan, err := handler.queries.GetScanByID(
				context.Background(),
				scanID,
			)

			if err == nil &&
				currentScan.Status == db.ScanStatusCancelled {
				cancel()
				return
			}
		}
	}
}

func normalizeImpact(
	impact string,
) db.ViolationImpact {
	switch strings.ToLower(impact) {
	case "critical":
		return db.ViolationImpactCritical
	case "serious":
		return db.ViolationImpactSerious
	case "moderate":
		return db.ViolationImpactModerate
	default:
		return db.ViolationImpactMinor
	}
}

func calculateScore(
	critical int32,
	serious int32,
	moderate int32,
	minor int32,
) int16 {
	score := 100 -
		int(critical)*15 -
		int(serious)*8 -
		int(moderate)*4 -
		int(minor)

	if score < 0 {
		score = 0
	}

	return int16(score)
}

func sanitizeError(err error) string {
	message := strings.TrimSpace(err.Error())

	if len(message) > 1000 {
		message = message[:1000]
	}

	return message
}
