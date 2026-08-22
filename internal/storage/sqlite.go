package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ikwukao/strata-log/internal/ingest"
	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS logs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	timestamp TEXT NOT NULL,
	level TEXT NOT NULL,
	service TEXT NOT NULL,
	message TEXT NOT NULL,
	fields TEXT
);

CREATE INDEX IF NOT EXISTS idx_logs_timestamp
	ON logs(timestamp);

CREATE INDEX IF NOT EXISTS idx_logs_level
	ON logs(level);

CREATE INDEX IF NOT EXISTS idx_logs_service
	ON logs(service);
`

// SQLiteWriter persists log entries in a SQLite database.
type SQLiteWriter struct {
	db *sql.DB

	// insertHook is used by tests to simulate an insertion failure.
	insertHook func(ingest.LogEntry) error
}

// NewSQLiteWriter opens the database and initializes its schema.
//
// SQLite WAL mode is enabled to improve concurrent read/write behavior.
func NewSQLiteWriter(path string) (*SQLiteWriter, error) {
	if path == "" {
		return nil, errors.New("storage path is required")
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	if _, err := db.Exec(`
		PRAGMA journal_mode = WAL;
		PRAGMA synchronous = NORMAL;
		PRAGMA foreign_keys = ON;
	`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("configure sqlite: %w", err)
	}

	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("initialize schema: %w", err)
	}

	return &SQLiteWriter{
		db: db,
	}, nil
}

// WriteBatch atomically persists a batch of log entries.
func (w *SQLiteWriter) WriteBatch(
	ctx context.Context,
	entries []ingest.LogEntry,
) error {
	if len(entries) == 0 {
		return nil
	}

	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO logs (
			timestamp,
			level,
			service,
			message,
			fields
		)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	for _, entry := range entries {
		fields := ""

		if entry.Fields != nil {
			data, err := json.Marshal(entry.Fields)
			if err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("marshal fields: %w", err)
			}

			fields = string(data)
		}

		if w.insertHook != nil {
			if err := w.insertHook(entry); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("insert hook: %w", err)
			}
		}
		_, err := stmt.ExecContext(
			ctx,
			entry.Timestamp.UTC().Format(time.RFC3339Nano),
			entry.Level,
			entry.Service,
			entry.Message,
			fields,
		)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("insert log entry: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// Close closes the SQLite database.
func (w *SQLiteWriter) Close() error {
	if w.db == nil {
		return nil
	}

	return w.db.Close()
}
