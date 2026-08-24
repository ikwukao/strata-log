package telemetry

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetricsCounters(t *testing.T) {
	metrics := NewMetrics()

	metrics.IncIngested()
	metrics.IncIngested()
	metrics.IncStored()
	metrics.IncErrors()

	ingested, stored, errors := metrics.Snapshot()

	if ingested != 2 {
		t.Fatalf("expected 2 ingested entries, got %d", ingested)
	}

	if stored != 1 {
		t.Fatalf("expected 1 stored entry, got %d", stored)
	}

	if errors != 1 {
		t.Fatalf("expected 1 error, got %d", errors)
	}
}

func TestMetricsHandler(t *testing.T) {
	metrics := NewMetrics()

	metrics.IncIngested()
	metrics.IncStored()
	metrics.IncErrors()

	req := httptest.NewRequest(
		http.MethodGet,
		"/metrics",
		nil,
	)

	rec := httptest.NewRecorder()

	metrics.Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d",
			rec.Code,
		)
	}

	body := rec.Body.String()

	expected := []string{
		"strata_log_ingested_total 1",
		"strata_log_stored_total 1",
		"strata_log_errors_total 1",
	}

	for _, value := range expected {
		if !strings.Contains(body, value) {
			t.Fatalf(
				"expected metrics output to contain %q\n%s",
				value,
				body,
			)
		}
	}
}
