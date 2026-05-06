package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"agent-cluster/internal/indexer"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx := context.Background()

	if len(os.Args) < 2 {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
			"ok":    false,
			"error": "missing command",
		})
		return 2
	}

	switch os.Args[1] {
	case "index":
		return runIndex(ctx, os.Args[2:])
	case "query":
		return runQuery(ctx, os.Args[2:])
	case "search":
		return runSearch(ctx, os.Args[2:])
	case "refs":
		return runRefs(ctx, os.Args[2:])
	default:
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
			"ok":    false,
			"error": fmt.Sprintf("unknown command: %s", os.Args[1]),
		})
		return 2
	}
}

func runIndex(ctx context.Context, args []string) int {
	start := time.Now()

	fs := flag.NewFlagSet("index", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	dbPath := fs.String("db", filepath.FromSlash(".agent-indexer/index.db"), "db path")
	if err := fs.Parse(args); err != nil {
		return writeErr(2, err)
	}
	rest := fs.Args()
	if len(rest) < 1 {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
			"ok":    false,
			"error": "missing <path>",
		})
		return 2
	}

	targetPath := rest[0]

	summary, err := indexer.Index(ctx, indexer.IndexParams{
		Target: targetPath,
		DBPath: *dbPath,
		Now:    time.Now,
	})
	if err != nil {
		return writeErr(1, err)
	}

	summary.DurationMS = time.Since(start).Milliseconds()
	_ = json.NewEncoder(os.Stdout).Encode(summary)
	return 0
}

func runQuery(ctx context.Context, args []string) int {
	fs := flag.NewFlagSet("query", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	dbPath := fs.String("db", filepath.FromSlash(".agent-indexer/index.db"), "db path")
	name := fs.String("name", "", "symbol name")
	limit := fs.Int("limit", 50, "max results")
	if err := fs.Parse(args); err != nil {
		return writeErr(2, err)
	}

	resp, err := indexer.QueryByName(ctx, *dbPath, *name, *limit)
	if err != nil {
		return writeErr(1, err)
	}
	_ = json.NewEncoder(os.Stdout).Encode(resp)
	return 0
}

func runSearch(ctx context.Context, args []string) int {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	dbPath := fs.String("db", filepath.FromSlash(".agent-indexer/index.db"), "db path")
	limit := fs.Int("limit", 20, "max results")
	if err := fs.Parse(args); err != nil {
		return writeErr(2, err)
	}
	rest := fs.Args()
	if len(rest) < 1 {
		return writeErr(2, fmt.Errorf("missing <text>"))
	}

	resp, err := indexer.SearchText(ctx, *dbPath, rest[0], *limit)
	if err != nil {
		return writeErr(1, err)
	}
	_ = json.NewEncoder(os.Stdout).Encode(resp)
	return 0
}

func runRefs(ctx context.Context, args []string) int {
	fs := flag.NewFlagSet("refs", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	dbPath := fs.String("db", filepath.FromSlash(".agent-indexer/index.db"), "db path")
	id := fs.Int64("id", 0, "symbol id")
	limit := fs.Int("limit", 200, "max results")
	if err := fs.Parse(args); err != nil {
		return writeErr(2, err)
	}

	resp, err := indexer.RefsBySymbolID(ctx, *dbPath, *id, *limit)
	if err != nil {
		return writeErr(1, err)
	}
	_ = json.NewEncoder(os.Stdout).Encode(resp)
	return 0
}

func writeErr(code int, err error) int {
	_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
		"ok":    false,
		"error": err.Error(),
	})
	if code == 0 {
		return 1
	}
	return code
}
