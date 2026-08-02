package queue

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

const AccessibilityScanTask = "accessibility:scan"

type AccessibilityScanPayload struct {
	ScanID string `json:"scan_id"`
	URL    string `json:"url"`
}

func NewAccessibilityScanTask(
	scanID string,
	url string,
	queueName string,
	timeout time.Duration,
) (*asynq.Task, []asynq.Option, error) {
	payload, err := json.Marshal(
		AccessibilityScanPayload{
			ScanID: scanID,
			URL:    url,
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"gagal membuat payload scan: %w",
			err,
		)
	}

	task := asynq.NewTask(
		AccessibilityScanTask,
		payload,
	)

	options := []asynq.Option{
		asynq.Queue(queueName),
		asynq.TaskID(TaskID(scanID)),
		asynq.Timeout(timeout),
		asynq.MaxRetry(2),
	}

	return task, options, nil
}

func TaskID(scanID string) string {
	return "scan:" + scanID
}
