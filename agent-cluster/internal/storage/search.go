package storage

import (
	"context"
	"database/sql"
	"fmt"
)

type SearchResult struct {
	ID        int64  `json:"id"`
	UID       string `json:"uid"`
	FilePath  string `json:"file_path"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Package   string `json:"package"`
	Signature string `json:"signature"`
	LineStart int    `json:"line_start"`
	LineEnd   int    `json:"line_end"`
}

func SearchSymbolsFTS(ctx context.Context, db *sql.DB, query string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := db.QueryContext(ctx, `
SELECT s.id, s.uid, s.file_path, s.name, s.kind, s.package, s.signature, s.line_start, s.line_end
FROM fts_symbols
JOIN symbols s ON s.id = fts_symbols.rowid
WHERE fts_symbols MATCH ?
ORDER BY s.id
LIMIT ?;
`, query, limit)
	if err != nil {
		return nil, fmt.Errorf("fts query: %w", err)
	}
	defer rows.Close()

	var out []SearchResult
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.ID, &r.UID, &r.FilePath, &r.Name, &r.Kind, &r.Package, &r.Signature, &r.LineStart, &r.LineEnd); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}
	return out, nil
}

func QuerySymbolsByName(ctx context.Context, db *sql.DB, name string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := db.QueryContext(ctx, `
SELECT id, uid, file_path, name, kind, package, signature, line_start, line_end
FROM symbols
WHERE name = ?
ORDER BY id
LIMIT ?;
`, name, limit)
	if err != nil {
		return nil, fmt.Errorf("query symbols: %w", err)
	}
	defer rows.Close()

	var out []SearchResult
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.ID, &r.UID, &r.FilePath, &r.Name, &r.Kind, &r.Package, &r.Signature, &r.LineStart, &r.LineEnd); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}
	return out, nil
}
