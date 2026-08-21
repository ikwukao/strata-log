package ingest

import (
	"errors"
	"strings"
	"time"
)

type LogEntry struct {
	Timestamp time.Time      `json:"timestamp"`
	Level     string         `json:"level"`
	Service   string         `json:"service"`
	Message   string         `json:"message"`
	Fields    map[string]any `json:"fields,omitempty"`
}

func (l LogEntry) Validate() error {
	if l.Timestamp.IsZero() {
		return errors.New("timestamp is required")
	}

	if strings.TrimSpace(l.Level) == "" {
		return errors.New("level is required")
	}

	if strings.TrimSpace(l.Service) == "" {
		return errors.New("service is required")
	}

	if strings.TrimSpace(l.Message) == "" {
		return errors.New("message is required")
	}

	return nil
}
