// Package query provides log retrieval and query services.
package query

import (
	"context"
	"errors"

	"github.com/ikwukao/strata-log/internal/storage"
)

// DefaultLimit is the default number of logs returned by a query.
const DefaultLimit = 100

// MaxLimit is the maximum number of logs returned by a query.
const MaxLimit = 1000

// Service provides read access to persisted logs.
type Service struct {
	reader storage.Reader
}

// NewService creates a query service backed by the supplied reader.
func NewService(reader storage.Reader) (*Service, error) {
	if reader == nil {
		return nil, errors.New("query reader is required")
	}

	return &Service{
		reader: reader,
	}, nil
}

// Query retrieves logs using the supplied options.
func (s *Service) Query(
	ctx context.Context,
	options storage.QueryOptions,
) ([]storage.LogRecord, error) {
	if options.Limit <= 0 {
		options.Limit = DefaultLimit
	}

	if options.Limit > MaxLimit {
		options.Limit = MaxLimit
	}

	return s.reader.QueryLogs(ctx, options)
}
