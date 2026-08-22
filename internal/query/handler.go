package query

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ikwukao/strata-log/internal/storage"
)

// Handler exposes persisted log queries over HTTP.
type Handler struct {
	service *Service
}

// NewHandler creates an HTTP query handler.
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// ServeHTTP handles GET /v1/logs requests.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	if h.service == nil {
		http.Error(
			w,
			"query service unavailable",
			http.StatusServiceUnavailable,
		)
		return
	}

	options := storage.QueryOptions{
		Level:   r.URL.Query().Get("level"),
		Service: r.URL.Query().Get("service"),
		Limit:   parseLimit(r.URL.Query().Get("limit")),
	}

	records, err := h.service.Query(r.Context(), options)
	if err != nil {
		http.Error(
			w,
			"failed to query logs",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(map[string]any{
		"logs":  records,
		"count": len(records),
	})
}

func parseLimit(value string) int {
	if value == "" {
		return DefaultLimit
	}

	limit, err := strconv.Atoi(value)
	if err != nil || limit <= 0 {
		return DefaultLimit
	}

	if limit > MaxLimit {
		return MaxLimit
	}

	return limit
}
