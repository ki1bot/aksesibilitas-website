package httpapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

func (server *Server) executeScan(
	ctx context.Context,
	scanID uuid.UUID,
) error {
	scannerURL := strings.TrimSpace(
		server.cfg.ScannerURL,
	)

	if scannerURL == "" {
		return fmt.Errorf(
			"SCANNER_URL belum dikonfigurasi",
		)
	}

	endpoint, err := url.JoinPath(
		scannerURL,
		"internal",
		"scans",
		scanID.String(),
	)
	if err != nil {
		return fmt.Errorf(
			"URL scanner tidak valid: %w",
			err,
		)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		nil,
	)
	if err != nil {
		return fmt.Errorf(
			"gagal membuat request scanner: %w",
			err,
		)
	}

	request.Header.Set(
		"Authorization",
		"Bearer "+server.cfg.ScannerToken,
	)

	response, err := server.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf(
			"scanner tidak dapat dihubungi: %w",
			err,
		)
	}
	defer response.Body.Close()

	responseBody, _ := io.ReadAll(
		io.LimitReader(
			response.Body,
			4096,
		),
	)

	if response.StatusCode >= 200 &&
		response.StatusCode < 300 {
		return nil
	}

	message := strings.TrimSpace(
		string(responseBody),
	)

	if message == "" {
		message = response.Status
	}

	return fmt.Errorf(
		"scanner mengembalikan status %d: %s",
		response.StatusCode,
		message,
	)
}
