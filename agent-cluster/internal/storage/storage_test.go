package storage

import (
	"context"
	"path/filepath"
	"testing"

	"agent-cluster/internal/parser"
)

func TestFTS_SearchSymbolsFTS(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	db, err := Open(ctx, dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	files := []parser.FileRecord{
		{
			Path:          "main.go",
			SizeBytes:     1,
			MTimeUnixMS:   0,
			ContentHash:   "h",
			Language:      "go",
			LastIndexedAt: 0,
		},
	}
	if err := UpsertFiles(ctx, db, files); err != nil {
		t.Fatalf("UpsertFiles: %v", err)
	}

	symbols := []parser.SymbolRecord{
		{
			UID:       "u1",
			FilePath:  "main.go",
			Name:      "HelloWorld",
			Kind:      "func",
			Package:   "main",
			Signature: "func HelloWorld()",
			Doc:       "Says hello to the world.",
			Body:      "func HelloWorld() {}",
			LineStart: 1,
			LineEnd:   1,
			UpdatedAt: 0,
		},
	}
	if err := ReplaceSymbols(ctx, db, symbols); err != nil {
		t.Fatalf("ReplaceSymbols: %v", err)
	}

	results, err := SearchSymbolsFTS(ctx, db, "HelloWorld", 10)
	if err != nil {
		t.Fatalf("SearchSymbolsFTS: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Name != "HelloWorld" {
		t.Fatalf("unexpected result name: %q", results[0].Name)
	}
}

