package ingest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	HealthHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if body := strings.TrimSpace(rec.Body.String()); body != `{"status":"ok"}` {
		t.Fatalf("unexpected response body: %s", body)
	}
}

func TestLogHandlerAcceptsValidLog(t *testing.T) {
	body := `{
		"timestamp": "2026-08-21T18:00:00Z",
		"level": "error",
		"service": "payment-api",
		"message": "payment processing failed",
		"fields": {
			"request_id": "req_123"
		}
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/logs",
		strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	LogHandler(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}
}

func TestLogHandlerRejectsInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/logs",
		strings.NewReader(`{"timestamp":`),
	)

	rec := httptest.NewRecorder()

	LogHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestLogHandlerRejectsInvalidLog(t *testing.T) {
	body := `{
		"timestamp": "2026-08-21T18:00:00Z",
		"level": "error",
		"service": "",
		"message": "payment processing failed"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/logs",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	LogHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestLogHandlerRejectsWrongMethod(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/logs",
		nil,
	)

	rec := httptest.NewRecorder()

	LogHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestLogEntryValidate(t *testing.T) {
	valid := LogEntry{
		Timestamp: time.Now(),
		Level:     "info",
		Service:   "api",
		Message:   "request completed",
	}

	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid log entry, got error: %v", err)
	}

	tests := []struct {
		name  string
		entry LogEntry
	}{
		{
			name: "missing timestamp",
			entry: LogEntry{
				Level:   "info",
				Service: "api",
				Message: "request completed",
			},
		},
		{
			name: "missing level",
			entry: LogEntry{
				Timestamp: time.Now(),
				Service:   "api",
				Message:   "request completed",
			},
		},
		{
			name: "missing service",
			entry: LogEntry{
				Timestamp: time.Now(),
				Level:     "info",
				Message:   "request completed",
			},
		},
		{
			name: "missing message",
			entry: LogEntry{
				Timestamp: time.Now(),
				Level:     "info",
				Service:   "api",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.entry.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
