package storage

import (
	"context"
	"database/sql"
	"fmt"

	"agent-cluster/internal/parser"
)

func UpsertFiles(ctx context.Context, db *sql.DB, files []parser.FileRecord) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
INSERT INTO files(path,size_bytes,mtime_unix_ms,content_hash,language,last_indexed_at)
VALUES(?,?,?,?,?,?)
ON CONFLICT(path) DO UPDATE SET
  size_bytes=excluded.size_bytes,
  mtime_unix_ms=excluded.mtime_unix_ms,
  content_hash=excluded.content_hash,
  language=excluded.language,
  last_indexed_at=excluded.last_indexed_at
`)
	if err != nil {
		return fmt.Errorf("prepare upsert files: %w", err)
	}
	defer stmt.Close()

	for _, f := range files {
		if _, err := stmt.ExecContext(ctx, f.Path, f.SizeBytes, f.MTimeUnixMS, f.ContentHash, f.Language, f.LastIndexedAt); err != nil {
			return fmt.Errorf("upsert file %s: %w", f.Path, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func ReplaceSymbols(ctx context.Context, db *sql.DB, symbols []parser.SymbolRecord) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM symbols;`); err != nil {
		return fmt.Errorf("delete symbols: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, `
INSERT INTO symbols(
  uid,file_path,name,kind,package,receiver,signature,doc,body,
  line_start,line_end,col_start,col_end,content_hash,updated_at
) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
`)
	if err != nil {
		return fmt.Errorf("prepare insert symbols: %w", err)
	}
	defer stmt.Close()

	for _, s := range symbols {
		if _, err := stmt.ExecContext(ctx,
			s.UID, s.FilePath, s.Name, s.Kind, s.Package, s.Receiver, s.Signature, s.Doc, s.Body,
			s.LineStart, s.LineEnd, s.ColStart, s.ColEnd, s.ContentHash, s.UpdatedAt,
		); err != nil {
			return fmt.Errorf("insert symbol %s: %w", s.UID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func ReplaceSymbolsForFiles(ctx context.Context, db *sql.DB, filePaths []string, symbols []parser.SymbolRecord) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	delStmt, err := tx.PrepareContext(ctx, `DELETE FROM symbols WHERE file_path = ?;`)
	if err != nil {
		return fmt.Errorf("prepare delete symbols by file: %w", err)
	}
	defer delStmt.Close()
	for _, fp := range filePaths {
		if _, err := delStmt.ExecContext(ctx, fp); err != nil {
			return fmt.Errorf("delete symbols for %s: %w", fp, err)
		}
	}

	insStmt, err := tx.PrepareContext(ctx, `
INSERT INTO symbols(
  uid,file_path,name,kind,package,receiver,signature,doc,body,
  line_start,line_end,col_start,col_end,content_hash,updated_at
) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
`)
	if err != nil {
		return fmt.Errorf("prepare insert symbols: %w", err)
	}
	defer insStmt.Close()

	for _, s := range symbols {
		if _, err := insStmt.ExecContext(ctx,
			s.UID, s.FilePath, s.Name, s.Kind, s.Package, s.Receiver, s.Signature, s.Doc, s.Body,
			s.LineStart, s.LineEnd, s.ColStart, s.ColEnd, s.ContentHash, s.UpdatedAt,
		); err != nil {
			return fmt.Errorf("insert symbol %s: %w", s.UID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
