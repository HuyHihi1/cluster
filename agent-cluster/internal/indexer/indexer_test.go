package indexer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIndex_IncrementalByHash(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tmp := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/m\n\ngo 1.24.3\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	src := `package main

// Hello says hi.
func Hello() string { return "hi" }
`
	if err := os.WriteFile(filepath.Join(tmp, "main.go"), []byte(src), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}

	dbPath := filepath.Join(tmp, "index.db")
	now := func() time.Time { return time.Unix(0, 0) }

	s1, err := Index(ctx, IndexParams{Target: tmp + "/...", DBPath: dbPath, Now: now})
	if err != nil {
		t.Fatalf("Index #1: %v", err)
	}
	if s1.UpdatedFiles == 0 {
		t.Fatalf("expected updated files > 0, got %d", s1.UpdatedFiles)
	}

	s2, err := Index(ctx, IndexParams{Target: tmp + "/...", DBPath: dbPath, Now: now})
	if err != nil {
		t.Fatalf("Index #2: %v", err)
	}
	if s2.UpdatedFiles != 0 {
		t.Fatalf("expected updated_files=0 on second run, got %d", s2.UpdatedFiles)
	}
	if s2.SkippedFiles != s2.IndexedFiles {
		t.Fatalf("expected all skipped on second run, got skipped=%d indexed=%d", s2.SkippedFiles, s2.IndexedFiles)
	}
}

func TestIndex_Refs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tmp := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/m\n\ngo 1.24.3\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "hello.go"), []byte("package main\n\nfunc Hello() {}\n"), 0o644); err != nil {
		t.Fatalf("write hello.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "main.go"), []byte("package main\n\nfunc main() { Hello() }\n"), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}

	dbPath := filepath.Join(tmp, "index.db")
	now := func() time.Time { return time.Unix(0, 0) }

	if _, err := Index(ctx, IndexParams{Target: tmp + "/...", DBPath: dbPath, Now: now}); err != nil {
		t.Fatalf("Index: %v", err)
	}

	q, err := QueryByName(ctx, dbPath, "Hello", 10)
	if err != nil {
		t.Fatalf("QueryByName: %v", err)
	}
	if len(q.Results) != 1 {
		t.Fatalf("expected 1 Hello symbol, got %d", len(q.Results))
	}
	symID := q.Results[0].ID

	r, err := RefsBySymbolID(ctx, dbPath, symID, 50)
	if err != nil {
		t.Fatalf("RefsBySymbolID: %v", err)
	}
	if len(r.Results) == 0 {
		t.Fatalf("expected refs to Hello, got 0")
	}
}
