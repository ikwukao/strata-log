package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ikwukao/strata-log/internal/batcher"
	"github.com/ikwukao/strata-log/internal/ingest"
	"github.com/ikwukao/strata-log/internal/pipeline"
	"github.com/ikwukao/strata-log/internal/query"
	"github.com/ikwukao/strata-log/internal/storage"
)

func TestLogIngestionAndQuery(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "strata-log.db")

	writer, err := storage.NewSQLiteWriter(dbPath)
	if err != nil {
		t.Fatalf("create storage: %v", err)
	}
	defer writer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b, err := batcher.New(
		ctx,
		writer,
		2,
		50*time.Millisecond,
		10,
		3,
		10*time.Millisecond,
	)
	if err != nil {
		t.Fatalf("create batcher: %v", err)
	}
	defer b.Close()

	processor, err := pipeline.NewLogProcessor(
		ctx,
		2,
		10,
		func(ctx context.Context, entry ingest.LogEntry) error {
			return b.Submit(entry)
		},
	)
	if err != nil {
		t.Fatalf("create processor: %v", err)
	}
	defer processor.Close()

	queryService, err := query.NewService(writer)
	if err != nil {
		t.Fatalf("create query service: %v", err)
	}

	ingestHandler := ingest.ProcessHandler(processor)
	queryHandler := query.NewHandler(queryService)

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/logs", func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		switch r.Method {
		case http.MethodPost:
			ingestHandler(w, r)
		case http.MethodGet:
			queryHandler.ServeHTTP(w, r)
		default:
			http.Error(
				w,
				"method not allowed",
				http.StatusMethodNotAllowed,
			)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	payload := map[string]any{
		"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		"level":     "info",
		"service":   "api",
		"message":   "integration test log",
		"fields": map[string]any{
			"environment": "test",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	resp, err := http.Post(
		server.URL+"/v1/logs",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("POST /v1/logs: %v", err)
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf(
			"expected successful ingestion, got HTTP %d",
			resp.StatusCode,
		)
	}

	// Give the asynchronous pipeline time to persist the batch.
	select {
	case count := <-b.Stored():
		if count != 1 {
			t.Fatalf("expected 1 stored entry, got %d", count)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for log persistence")
	}

	resp, err = http.Get(server.URL + "/v1/logs?service=api&limit=10")
	if err != nil {
		t.Fatalf("GET /v1/logs: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(
			"expected HTTP 200 from query, got %d",
			resp.StatusCode,
		)
	}

	var result struct {
		Logs  []storage.LogRecord `json:"logs"`
		Count int                 `json:"count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode query response: %v", err)
	}

	if result.Count != 1 {
		t.Fatalf("expected 1 queried log, got %d", result.Count)
	}

	if len(result.Logs) != 1 {
		t.Fatalf("expected 1 log record, got %d", len(result.Logs))
	}

	record := result.Logs[0]

	if record.Level != "info" {
		t.Fatalf("expected level info, got %q", record.Level)
	}

	if record.Service != "api" {
		t.Fatalf("expected service api, got %q", record.Service)
	}

	if record.Message != "integration test log" {
		t.Fatalf(
			"expected message %q, got %q",
			"integration test log",
			record.Message,
		)
	}

	if record.Fields["environment"] != "test" {
		t.Fatalf(
			"expected environment=test, got %v",
			record.Fields["environment"],
		)
	}

	_ = os.Remove(dbPath)
}
