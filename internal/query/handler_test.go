package query

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ikwukao/strata-log/internal/storage"
)

func TestHandlerReturnsLogs(t *testing.T) {
	reader := &mockReader{
		records: []storage.LogRecord{
			{
				ID:        1,
				Timestamp: "2026-08-22T20:00:00Z",
				Level:     "error",
				Service:   "payment-api",
				Message:   "database unavailable",
				Fields: map[string]any{
					"request_id": "req-1",
				},
			},
		},
	}

	service, err := NewService(reader)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/logs",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	if reader.options.Limit != DefaultLimit {
		t.Fatalf(
			"expected default limit %d, got %d",
			DefaultLimit,
			reader.options.Limit,
		)
	}

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf(
			"expected application/json, got %q",
			rec.Header().Get("Content-Type"),
		)
	}
}

func TestHandlerPassesFilters(t *testing.T) {
	reader := &mockReader{}

	service, err := NewService(reader)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/logs?level=error&service=payment-api&limit=25",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	if reader.options.Level != "error" {
		t.Fatalf(
			"expected level error, got %q",
			reader.options.Level,
		)
	}

	if reader.options.Service != "payment-api" {
		t.Fatalf(
			"expected payment-api, got %q",
			reader.options.Service,
		)
	}

	if reader.options.Limit != 25 {
		t.Fatalf(
			"expected limit 25, got %d",
			reader.options.Limit,
		)
	}
}

func TestHandlerRejectsNonGET(t *testing.T) {
	reader := &mockReader{}

	service, err := NewService(reader)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/logs",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusMethodNotAllowed,
			rec.Code,
		)
	}
}

func TestHandlerRejectsUnavailableService(t *testing.T) {
	handler := NewHandler(nil)

	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/logs",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusServiceUnavailable,
			rec.Code,
		)
	}
}

func TestHandlerUsesRequestContext(t *testing.T) {
	reader := &contextReader{}

	service, err := NewService(reader)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/logs",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}

	if !reader.called {
		t.Fatal("expected request context to reach reader")
	}
}

type contextReader struct {
	called bool
}

func (r *contextReader) QueryLogs(
	ctx context.Context,
	_ storage.QueryOptions,
) ([]storage.LogRecord, error) {
	r.called = ctx != nil

	return nil, nil
}
