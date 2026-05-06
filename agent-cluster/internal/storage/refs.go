package storage

import (
	"context"
	"database/sql"
	"fmt"

	"agent-cluster/internal/parser"
)

type RefRow struct {
	SymbolID int64  `json:"symbol_id"`
	FilePath string `json:"file_path"`
	Line     int    `json:"line"`
	Col      int    `json:"col"`
}

func ReplaceRefsForFiles(ctx context.Context, db *sql.DB, filePaths []string, refs []parser.RefRecord) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	delStmt, err := tx.PrepareContext(ctx, `DELETE FROM refs WHERE file_path = ?;`)
	if err != nil {
		return fmt.Errorf("prepare delete refs: %w", err)
	}
	defer delStmt.Close()
	for _, fp := range filePaths {
		if _, err := delStmt.ExecContext(ctx, fp); err != nil {
			return fmt.Errorf("delete refs for %s: %w", fp, err)
		}
	}

	insStmt, err := tx.PrepareContext(ctx, `
INSERT INTO refs(symbol_id, file_path, line, col)
SELECT id, ?, ?, ? FROM symbols WHERE uid = ?;
`)
	if err != nil {
		return fmt.Errorf("prepare insert ref: %w", err)
	}
	defer insStmt.Close()

	for _, r := range refs {
		if r.ToUID == "" {
			continue
		}
		if _, err := insStmt.ExecContext(ctx, r.FilePath, r.Line, r.Col, r.ToUID); err != nil {
			return fmt.Errorf("insert ref to %s: %w", r.ToUID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func ListRefsBySymbolID(ctx context.Context, db *sql.DB, symbolID int64, limit int) ([]RefRow, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := db.QueryContext(ctx, `
SELECT symbol_id, file_path, line, col
FROM refs
WHERE symbol_id = ?
ORDER BY file_path, line, col
LIMIT ?;
`, symbolID, limit)
	if err != nil {
		return nil, fmt.Errorf("list refs: %w", err)
	}
	defer rows.Close()

	var out []RefRow
	for rows.Next() {
		var r RefRow
		if err := rows.Scan(&r.SymbolID, &r.FilePath, &r.Line, &r.Col); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}
	return out, nil
}

