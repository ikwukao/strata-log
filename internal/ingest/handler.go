// Package ingest provides HTTP ingestion handlers and data models
// for receiving structured log entries.
package ingest

import (
	"encoding/json"
	"net/http"
)

// maxRequestBodySize limits individual ingestion requests to 1 MiB.
const maxRequestBodySize = 1 << 20

// HealthHandler reports the health of the Strata-Log HTTP service.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// LogHandler accepts structured log entries through POST /v1/logs.
//
// Requests are limited to 1 MiB and unknown JSON fields are rejected.
// Valid entries are acknowledged with HTTP 202 Accepted.
func LogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	defer r.Body.Close()

	var entry LogEntry

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&entry); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := entry.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "accepted",
	})
}
