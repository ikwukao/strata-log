// Package telemetry provides runtime metrics for Strata-Log.
package telemetry

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

// Metrics stores application-level counters.
//
// Counters use atomic operations so they can be updated safely
// by concurrent HTTP handlers and background workers.
type Metrics struct {
	ingested uint64
	stored   uint64
	errors   uint64
}

// NewMetrics creates an initialized metrics collector.
func NewMetrics() *Metrics {
	return &Metrics{}
}

// IncIngested increments the number of accepted log entries.
func (m *Metrics) IncIngested() {
	atomic.AddUint64(&m.ingested, 1)
}

// IncStored increments the number of persisted log entries.
func (m *Metrics) IncStored() {
	atomic.AddUint64(&m.stored, 1)
}

// IncErrors increments the number of processing or storage errors.
func (m *Metrics) IncErrors() {
	atomic.AddUint64(&m.errors, 1)
}

// Snapshot returns the current counter values.
func (m *Metrics) Snapshot() (ingested, stored, errors uint64) {
	return atomic.LoadUint64(&m.ingested),
		atomic.LoadUint64(&m.stored),
		atomic.LoadUint64(&m.errors)
}

// ServeHTTP exposes metrics using the Prometheus text exposition format.
func (m *Metrics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	m.Handler(w, r)
}

// Handler exposes metrics using the Prometheus text exposition format.
func (m *Metrics) Handler(w http.ResponseWriter, _ *http.Request) {
	ingested, stored, errors := m.Snapshot()

	w.Header().Set(
		"Content-Type",
		"text/plain; version=0.0.4",
	)

	_, _ = fmt.Fprintf(
		w,
		"# HELP strata_log_ingested_total Total accepted log entries.\n"+
			"# TYPE strata_log_ingested_total counter\n"+
			"strata_log_ingested_total %d\n"+
			"# HELP strata_log_stored_total Total persisted log entries.\n"+
			"# TYPE strata_log_stored_total counter\n"+
			"strata_log_stored_total %d\n"+
			"# HELP strata_log_errors_total Total processing and storage errors.\n"+
			"# TYPE strata_log_errors_total counter\n"+
			"strata_log_errors_total %d\n",
		ingested,
		stored,
		errors,
	)
}
