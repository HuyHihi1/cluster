package indexer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"agent-cluster/internal/parser"
	"agent-cluster/internal/storage"
	"agent-cluster/internal/util"
)

type IndexParams struct {
	Target string
	DBPath string
	Now    func() time.Time
}

type IndexSummary struct {
	OK           bool   `json:"ok"`
	DBPath       string `json:"db_path"`
	IndexedFiles int    `json:"indexed_files"`
	UpdatedFiles int    `json:"updated_files"`
	SkippedFiles int    `json:"skipped_files"`
	DurationMS   int64  `json:"duration_ms"`
}

func Index(ctx context.Context, p IndexParams) (IndexSummary, error) {
	if p.Now == nil {
		return IndexSummary{}, errors.New("missing Now")
	}
	if p.Target == "" {
		return IndexSummary{}, errors.New("missing Target")
	}
	if p.DBPath == "" {
		return IndexSummary{}, errors.New("missing DBPath")
	}

	if err := util.CheckContext(ctx); err != nil {
		return IndexSummary{}, err
	}
	if err := os.MkdirAll(filepath.Dir(p.DBPath), 0o755); err != nil {
		return IndexSummary{}, fmt.Errorf("mkdir db dir: %w", err)
	}

	db, err := storage.Open(ctx, p.DBPath)
	if err != nil {
		return IndexSummary{}, err
	}
	defer db.Close()

	nowMS := p.Now().UnixMilli()

	candidatePaths, err := collectGoFiles(p.Target)
	if err != nil {
		return IndexSummary{}, err
	}
	existing, err := storage.GetFileContentHashesByPath(ctx, db, candidatePaths)
	if err != nil {
		return IndexSummary{}, err
	}

	parseResult, err := parser.ParseWorkspaceWithOptions(ctx, p.Target, nowMS, parser.ParseOptions{
		ExistingContentHashByPath: existing,
	})
	if err != nil {
		return IndexSummary{}, err
	}

	currentPaths := make([]string, 0, len(parseResult.Files))
	changedPaths := make([]string, 0, len(parseResult.Files))
	for _, f := range parseResult.Files {
		fp := filepath.Clean(f.Path)
		currentPaths = append(currentPaths, fp)
		if prev, ok := existing[fp]; !ok || prev != f.ContentHash {
			changedPaths = append(changedPaths, fp)
		}
	}
	sort.Strings(currentPaths)
	sort.Strings(changedPaths)

	prevPaths, err := storage.ListIndexedFilePaths(ctx, db)
	if err != nil {
		return IndexSummary{}, err
	}
	removedPaths := diffSorted(prevPaths, currentPaths)

	if err := storage.DeleteFiles(ctx, db, removedPaths); err != nil {
		return IndexSummary{}, err
	}
	if err := storage.UpsertFiles(ctx, db, parseResult.Files); err != nil {
		return IndexSummary{}, err
	}
	if err := storage.ReplaceSymbolsForFiles(ctx, db, changedPaths, parseResult.Symbols); err != nil {
		return IndexSummary{}, err
	}
	if err := storage.ReplaceRefsForFiles(ctx, db, changedPaths, parseResult.Refs); err != nil {
		return IndexSummary{}, err
	}

	return IndexSummary{
		OK:           true,
		DBPath:       p.DBPath,
		IndexedFiles: len(parseResult.Files),
		UpdatedFiles: len(changedPaths),
		SkippedFiles: len(parseResult.Files) - len(changedPaths),
	}, nil
}

func collectGoFiles(target string) ([]string, error) {
	base := target
	if strings.HasSuffix(base, "/...") {
		base = strings.TrimSuffix(base, "/...")
		if base == "" {
			base = "."
		}
	}

	st, err := os.Stat(base)
	if err == nil && !st.IsDir() {
		if strings.HasSuffix(base, ".go") {
			abs, err := filepath.Abs(base)
			if err != nil {
				return nil, fmt.Errorf("abs %s: %w", base, err)
			}
			return []string{filepath.Clean(abs)}, nil
		}
		return nil, nil
	}

	root := base
	if root == "" {
		root = "."
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("abs %s: %w", root, err)
	}

	var out []string
	err = filepath.WalkDir(absRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		name := d.Name()
		if d.IsDir() {
			if name == "vendor" || name == ".agent-indexer" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(name, ".go") {
			out = append(out, filepath.Clean(path))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", absRoot, err)
	}
	sort.Strings(out)
	return out, nil
}

func diffSorted(a, b []string) []string {
	// Returns elements in a that are not in b; both must be sorted.
	var out []string
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i] == b[j] {
			i++
			j++
			continue
		}
		if a[i] < b[j] {
			out = append(out, a[i])
			i++
			continue
		}
		j++
	}
	for i < len(a) {
		out = append(out, a[i])
		i++
	}
	return out
}
