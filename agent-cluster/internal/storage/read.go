package storage

import (
	"context"
	"database/sql"
	"fmt"
)

func GetFileContentHashesByPath(ctx context.Context, db *sql.DB, paths []string) (map[string]string, error) {
	out := make(map[string]string, len(paths))
	if len(paths) == 0 {
		return out, nil
	}

	// Simple approach: query per path (Phase 2 size is small; optimize later if needed).
	stmt, err := db.PrepareContext(ctx, `SELECT content_hash FROM files WHERE path = ?;`)
	if err != nil {
		return nil, fmt.Errorf("prepare select file hash: %w", err)
	}
	defer stmt.Close()

	for _, p := range paths {
		var h string
		err := stmt.QueryRowContext(ctx, p).Scan(&h)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("select file hash %s: %w", p, err)
		}
		out[p] = h
	}
	return out, nil
}

func ListIndexedFilePaths(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, `SELECT path FROM files ORDER BY path;`)
	if err != nil {
		return nil, fmt.Errorf("list file paths: %w", err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, fmt.Errorf("scan file path: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}
	return out, nil
}

func DeleteFiles(ctx context.Context, db *sql.DB, paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `DELETE FROM files WHERE path = ?;`)
	if err != nil {
		return fmt.Errorf("prepare delete file: %w", err)
	}
	defer stmt.Close()

	for _, p := range paths {
		if _, err := stmt.ExecContext(ctx, p); err != nil {
			return fmt.Errorf("delete file %s: %w", p, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

