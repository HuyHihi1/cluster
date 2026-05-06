package storage

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func Open(ctx context.Context, dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys=ON;"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("pragma foreign_keys: %w", err)
	}
	if _, err := db.ExecContext(ctx, "PRAGMA trusted_schema=ON;"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("pragma trusted_schema: %w", err)
	}

	if _, err := db.ExecContext(ctx, schemaSQL); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	return db, nil
}
