// Package ingest provides HTTP ingestion handlers and data models
// for receiving structured log entries.
package ingest

import (
	"encoding/json"
	"errors"
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

// ProcessHandler returns an HTTP handler that validates incoming log entries
// and submits them to the configured processor.
func ProcessHandler(processor Processor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if processor == nil {
			http.Error(
				w,
				"processor unavailable",
				http.StatusServiceUnavailable,
			)
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

		if err := processor.Submit(entry); err != nil {
			switch {
			case errors.Is(err, ErrProcessorClosed):
				http.Error(
					w,
					"ingestion service unavailable",
					http.StatusServiceUnavailable,
				)

			case errors.Is(err, ErrProcessorFull):
				http.Error(
					w,
					"ingestion buffer full",
					http.StatusServiceUnavailable,
				)

			default:
				http.Error(
					w,
					"failed to accept log entry",
					http.StatusInternalServerError,
				)
			}

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)

		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "accepted",
		})
	}
}
