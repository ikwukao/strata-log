package ingest

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type mockProcessor struct {
	submitted []LogEntry
	err       error
}

func (m *mockProcessor) Submit(entry LogEntry) error {
	if m.err != nil {
		return m.err
	}

	m.submitted = append(m.submitted, entry)

	return nil
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/healthz",
		nil,
	)

	rec := httptest.NewRecorder()

	HealthHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	if contentType := rec.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf(
			"expected content type application/json, got %q",
			contentType,
		)
	}

	if body := strings.TrimSpace(rec.Body.String()); body != `{"status":"ok"}` {
		t.Fatalf(
			"unexpected response body: %s",
			body,
		)
	}
}

func TestProcessHandlerAcceptsValidLog(t *testing.T) {
	processor := &mockProcessor{}

	body := `{
		"timestamp": "2026-08-21T18:00:00Z",
		"level": "error",
		"service": "payment-api",
		"message": "payment processing failed",
		"fields": {
			"request_id": "req_123",
			"environment": "production"
		}
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/logs",
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	handler := ProcessHandler(processor)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusAccepted,
			rec.Code,
		)
	}

	if contentType := rec.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf(
			"expected content type application/json, got %q",
			contentType,
		)
	}

	if len(processor.submitted) != 1 {
		t.Fatalf(
			"expected 1 submitted log, got %d",
			len(processor.submitted),
		)
	}

	entry := processor.submitted[0]

	if entry.Level != "error" {
		t.Fatalf(
			"expected level error, got %q",
			entry.Level,
		)
	}

	if entry.Service != "payment-api" {
		t.Fatalf(
			"expected service payment-api, got %q",
			entry.Service,
		)
	}

	if entry.Message != "payment processing failed" {
		t.Fatalf(
			"unexpected message: %q",
			entry.Message,
		)
	}

	if entry.Timestamp.IsZero() {
		t.Fatal("expected timestamp to be populated")
	}

	if entry.Fields["request_id"] != "req_123" {
		t.Fatalf(
			"expected request_id req_123, got %v",
			entry.Fields["request_id"],
		)
	}

	if entry.Fields["environment"] != "production" {
		t.Fatalf(
			"expected environment production, got %v",
			entry.Fields["environment"],
		)
	}
}

func TestProcessHandlerRejectsInvalidJSON(t *testing.T) {
	processor := &mockProcessor{}

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/logs",
		strings.NewReader(`{"timestamp":`),
	)

	rec := httptest.NewRecorder()

	handler := ProcessHandler(processor)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}

	if len(processor.submitted) != 0 {
		t.Fatal("invalid JSON should not be submitted")
	}
}

func TestProcessHandlerRejectsUnknownFields(t *testing.T) {
	processor := &mockProcessor{}

	body := `{
		"timestamp": "2026-08-21T18:00:00Z",
		"level": "info",
		"service": "api",
		"message": "request completed",
		"unknown": "field"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/logs",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	handler := ProcessHandler(processor)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}

	if len(processor.submitted) != 0 {
		t.Fatal("request with unknown fields should not be submitted")
	}
}

func TestProcessHandlerRejectsMissingTimestamp(t *testing.T) {
	processor := &mockProcessor{}

	body := `{
		"level": "info",
		"service": "api",
		"message": "request completed"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/logs",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	handler := ProcessHandler(processor)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}

	if !strings.Contains(rec.Body.String(), "timestamp is required") {
		t.Fatalf(
			"expected timestamp validation error, got %q",
			rec.Body.String(),
		)
	}

	if len(processor.submitted) != 0 {
		t.Fatal("invalid log should not be submitted")
	}
}

func TestProcessHandlerRejectsInvalidLog(t *testing.T) {
	processor := &mockProcessor{}

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

	handler := ProcessHandler(processor)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}

	if len(processor.submitted) != 0 {
		t.Fatal("invalid log should not be submitted")
	}
}

func TestProcessHandlerRejectsMissingLevel(t *testing.T) {
	processor := &mockProcessor{}

	body := `{
		"timestamp": "2026-08-21T18:00:00Z",
		"service": "api",
		"message": "request completed"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/logs",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	handler := ProcessHandler(processor)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestProcessHandlerRejectsMissingService(t *testing.T) {
	processor := &mockProcessor{}

	body := `{
		"timestamp": "2026-08-21T18:00:00Z",
		"level": "info",
		"message": "request completed"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/logs",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	handler := ProcessHandler(processor)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestProcessHandlerRejectsMissingMessage(t *testing.T) {
	processor := &mockProcessor{}

	body := `{
		"timestamp": "2026-08-21T18:00:00Z",
		"level": "info",
		"service": "api"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/logs",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	handler := ProcessHandler(processor)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestProcessHandlerRejectsWrongMethod(t *testing.T) {
	processor := &mockProcessor{}

	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/logs",
		nil,
	)

	rec := httptest.NewRecorder()

	handler := ProcessHandler(processor)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusMethodNotAllowed,
			rec.Code,
		)
	}
}

func TestProcessHandlerRejectsNilProcessor(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/logs",
		strings.NewReader(`{}`),
	)

	rec := httptest.NewRecorder()

	handler := ProcessHandler(nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusServiceUnavailable,
			rec.Code,
		)
	}
}

func TestProcessHandlerProcessorClosed(t *testing.T) {
	processor := &mockProcessor{
		err: ErrProcessorClosed,
	}

	body := `{
		"timestamp": "2026-08-21T18:00:00Z",
		"level": "info",
		"service": "api",
		"message": "request completed"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/logs",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	handler := ProcessHandler(processor)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusServiceUnavailable,
			rec.Code,
		)
	}
}

func TestProcessHandlerProcessorFull(t *testing.T) {
	processor := &mockProcessor{
		err: ErrProcessorFull,
	}

	body := `{
		"timestamp": "2026-08-21T18:00:00Z",
		"level": "info",
		"service": "api",
		"message": "request completed"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/logs",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	handler := ProcessHandler(processor)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusServiceUnavailable,
			rec.Code,
		)
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
		t.Fatalf(
			"expected valid log entry, got error: %v",
			err,
		)
	}
}

func TestLogEntryValidateRequiresTimestamp(t *testing.T) {
	entry := LogEntry{
		Level:   "info",
		Service: "api",
		Message: "request completed",
	}

	err := entry.Validate()

	if err == nil {
		t.Fatal("expected timestamp validation error")
	}

	if err.Error() != "timestamp is required" {
		t.Fatalf(
			"expected timestamp is required, got %q",
			err.Error(),
		)
	}
}

func TestLogEntryValidateRequiresLevel(t *testing.T) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Service:   "api",
		Message:   "request completed",
	}

	err := entry.Validate()

	if err == nil {
		t.Fatal("expected level validation error")
	}
}

func TestLogEntryValidateRequiresService(t *testing.T) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "request completed",
	}

	err := entry.Validate()

	if err == nil {
		t.Fatal("expected service validation error")
	}
}

func TestLogEntryValidateRequiresMessage(t *testing.T) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     "info",
		Service:   "api",
	}

	err := entry.Validate()

	if err == nil {
		t.Fatal("expected message validation error")
	}
}

func TestProcessorErrors(t *testing.T) {
	if !errors.Is(ErrProcessorClosed, ErrProcessorClosed) {
		t.Fatal("expected ErrProcessorClosed to be comparable")
	}

	if !errors.Is(ErrProcessorFull, ErrProcessorFull) {
		t.Fatal("expected ErrProcessorFull to be comparable")
	}
}
