package queue

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

const TypeAccessibilityScan = "accessibility:scan"

type AccessibilityScanPayload struct {
	ScanID string `json:"scan_id"`
	URL    string `json:"url"`
}

func NewAccessibilityScanTask(
	scanID string,
	targetURL string,
	queueName string,
	timeout time.Duration,
) (*asynq.Task, error) {
	payload, err := json.Marshal(AccessibilityScanPayload{
		ScanID: scanID,
		URL:    targetURL,
	})
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(
		TypeAccessibilityScan,
		payload,
		asynq.Queue(queueName),
		asynq.MaxRetry(2),
		asynq.Timeout(timeout),
	), nil
}
