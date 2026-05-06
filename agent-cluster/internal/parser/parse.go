package parser

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"agent-cluster/internal/util"
	"golang.org/x/tools/go/packages"
)

func ParseWorkspace(ctx context.Context, pattern string, nowMS int64) (ParseResult, error) {
	return ParseWorkspaceWithOptions(ctx, pattern, nowMS, ParseOptions{})
}

type ParseOptions struct {
	ExistingContentHashByPath map[string]string
}

func ParseWorkspaceWithOptions(ctx context.Context, pattern string, nowMS int64, opt ParseOptions) (ParseResult, error) {
	loadDir, loadPattern, err := normalizePackagesLoad(pattern)
	if err != nil {
		return ParseResult{}, err
	}

	cfg := &packages.Config{
		Context: ctx,
		Dir:     loadDir,
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedModule,
	}

	pkgs, err := packages.Load(cfg, loadPattern)
	if err != nil {
		return ParseResult{}, fmt.Errorf("packages.Load: %w", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return ParseResult{}, fmt.Errorf("packages.Load: packages contain errors")
	}

	fileByPath := make(map[string]FileRecord)
	var symbols []SymbolRecord
	var refs []RefRecord
	existing := opt.ExistingContentHashByPath
	defUIDByObj := make(map[types.Object]string)

	for _, pkg := range pkgs {
		pkgPath := pkg.PkgPath
		for i, f := range pkg.Syntax {
			if i >= len(pkg.CompiledGoFiles) {
				continue
			}
			fp := pkg.CompiledGoFiles[i]

			if err := util.CheckContext(ctx); err != nil {
				return ParseResult{}, err
			}

			fr, err := statAndHashFile(ctx, fp, nowMS)
			if err != nil {
				return ParseResult{}, err
			}
			fr.Language = "go"
			fileByPath[fp] = fr

			if existing != nil {
				if prev, ok := existing[filepath.Clean(fp)]; ok && prev == fr.ContentHash {
					continue
				}
			}

			syms, err := extractSymbols(pkgPath, pkg.Types, pkg.TypesInfo, pkg.Fset, f, fp, nowMS, defUIDByObj)
			if err != nil {
				return ParseResult{}, err
			}
			symbols = append(symbols, syms...)

			r, err := extractRefs(pkg.TypesInfo, pkg.Fset, f, fp, defUIDByObj)
			if err != nil {
				return ParseResult{}, err
			}
			refs = append(refs, r...)
		}
	}

	files := make([]FileRecord, 0, len(fileByPath))
	for _, fr := range fileByPath {
		files = append(files, fr)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	sort.Slice(symbols, func(i, j int) bool {
		if symbols[i].FilePath == symbols[j].FilePath {
			if symbols[i].LineStart == symbols[j].LineStart {
				return symbols[i].Name < symbols[j].Name
			}
			return symbols[i].LineStart < symbols[j].LineStart
		}
		return symbols[i].FilePath < symbols[j].FilePath
	})

	sort.Slice(refs, func(i, j int) bool {
		if refs[i].FilePath == refs[j].FilePath {
			if refs[i].Line == refs[j].Line {
				return refs[i].Col < refs[j].Col
			}
			return refs[i].Line < refs[j].Line
		}
		return refs[i].FilePath < refs[j].FilePath
	})

	return ParseResult{Files: files, Symbols: symbols, Refs: refs}, nil
}

func normalizePackagesLoad(pattern string) (dir string, pat string, err error) {
	p := strings.TrimSpace(pattern)
	if p == "" {
		return "", "", fmt.Errorf("empty pattern")
	}

	// Support `indexer index /abs/path/...` by setting Dir to the base path and loading "./...".
	if strings.HasSuffix(p, "/...") && (filepath.IsAbs(p) || strings.HasPrefix(p, "."+string(filepath.Separator))) {
		base := strings.TrimSuffix(p, "/...")
		if base == "" {
			base = "."
		}
		abs, err := filepath.Abs(base)
		if err != nil {
			return "", "", fmt.Errorf("abs %s: %w", base, err)
		}
		return abs, "./...", nil
	}

	// Default: use current working directory.
	return "", p, nil
}

func statAndHashFile(ctx context.Context, path string, nowMS int64) (FileRecord, error) {
	if err := util.CheckContext(ctx); err != nil {
		return FileRecord{}, err
	}

	st, err := os.Stat(path)
	if err != nil {
		return FileRecord{}, fmt.Errorf("stat %s: %w", path, err)
	}
	f, err := os.Open(path)
	if err != nil {
		return FileRecord{}, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	h := sha256.New()
	r := util.NewContextReader(ctx, f)
	if _, err := io.Copy(h, r); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return FileRecord{}, err
		}
		return FileRecord{}, fmt.Errorf("hash %s: %w", path, err)
	}
	sum := hex.EncodeToString(h.Sum(nil))

	return FileRecord{
		Path:          filepath.Clean(path),
		SizeBytes:     st.Size(),
		MTimeUnixMS:   st.ModTime().UnixMilli(),
		ContentHash:   sum,
		LastIndexedAt: nowMS,
	}, nil
}

func extractSymbols(pkgPath string, pkgTypes *types.Package, info *types.Info, fset *token.FileSet, file *ast.File, filePath string, nowMS int64, defUIDByObj map[types.Object]string) ([]SymbolRecord, error) {
	var out []SymbolRecord

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			s, ok := info.Defs[d.Name].(*types.Func)
			if !ok || s == nil {
				continue
			}
			rec := ""
			if d.Recv != nil && len(d.Recv.List) > 0 {
				rec = exprString(fset, d.Recv.List[0].Type)
			}
			sig := ""
			if ts := s.Type(); ts != nil {
				sig = types.TypeString(ts, qualifierFor(pkgTypes))
			}
			doc := strings.TrimSpace(docText(d.Doc))
			body := ""
			if d.Body != nil {
				body = strings.TrimSpace(nodeString(fset, d.Body))
			}
			ls, le, cs, ce := spanForNode(fset, d)
			uid := stableUID(filePath, "func", s.Name(), ls, le, cs, ce)
			if defUIDByObj != nil {
				defUIDByObj[s] = uid
			}
			out = append(out, SymbolRecord{
				UID:       uid,
				FilePath:  filepath.Clean(filePath),
				Name:      s.Name(),
				Kind:      kindFunc(rec),
				Package:   pkgPath,
				Receiver:  rec,
				Signature: sig,
				Doc:       doc,
				Body:      body,
				LineStart: ls,
				LineEnd:   le,
				ColStart:  cs,
				ColEnd:    ce,
				UpdatedAt: nowMS,
			})

		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					obj := info.Defs[s.Name]
					if obj == nil {
						continue
					}
					doc := strings.TrimSpace(docText(d.Doc))
					if s.Doc != nil {
						doc = strings.TrimSpace(docText(s.Doc))
					}
					sig := ""
					if t := obj.Type(); t != nil {
						sig = types.TypeString(t, qualifierFor(pkgTypes))
					}
					ls, le, cs, ce := spanForNode(fset, s)
					uid := stableUID(filePath, "type", obj.Name(), ls, le, cs, ce)
					if defUIDByObj != nil {
						defUIDByObj[obj] = uid
					}
					out = append(out, SymbolRecord{
						UID:       uid,
						FilePath:  filepath.Clean(filePath),
						Name:      obj.Name(),
						Kind:      "type",
						Package:   pkgPath,
						Signature: sig,
						Doc:       doc,
						LineStart: ls,
						LineEnd:   le,
						ColStart:  cs,
						ColEnd:    ce,
						UpdatedAt: nowMS,
					})
				case *ast.ValueSpec:
					for _, n := range s.Names {
						obj := info.Defs[n]
						if obj == nil {
							continue
						}
						doc := strings.TrimSpace(docText(d.Doc))
						if s.Doc != nil {
							doc = strings.TrimSpace(docText(s.Doc))
						}
						sig := ""
						if t := obj.Type(); t != nil {
							sig = types.TypeString(t, qualifierFor(pkgTypes))
						}
						ls, le, cs, ce := spanForNode(fset, n)
						uid := stableUID(filePath, strings.ToLower(d.Tok.String()), obj.Name(), ls, le, cs, ce)
						if defUIDByObj != nil {
							defUIDByObj[obj] = uid
						}
						out = append(out, SymbolRecord{
							UID:       uid,
							FilePath:  filepath.Clean(filePath),
							Name:      obj.Name(),
							Kind:      strings.ToLower(d.Tok.String()),
							Package:   pkgPath,
							Signature: sig,
							Doc:       doc,
							LineStart: ls,
							LineEnd:   le,
							ColStart:  cs,
							ColEnd:    ce,
							UpdatedAt: nowMS,
						})
					}
				}
			}
		}
	}

	return out, nil
}

func qualifierFor(pkgTypes *types.Package) func(*types.Package) string {
	return func(p *types.Package) string {
		if pkgTypes != nil && p == pkgTypes {
			return ""
		}
		if p == nil {
			return ""
		}
		return p.Name()
	}
}

func kindFunc(receiver string) string {
	if receiver != "" {
		return "method"
	}
	return "func"
}

func docText(cg *ast.CommentGroup) string {
	if cg == nil {
		return ""
	}
	return cg.Text()
}

func stableUID(parts ...any) string {
	h := sha256.New()
	for _, p := range parts {
		_, _ = io.WriteString(h, fmt.Sprint(p))
		_, _ = io.WriteString(h, "|")
	}
	return hex.EncodeToString(h.Sum(nil))
}

func spanForNode(fset *token.FileSet, n ast.Node) (lineStart, lineEnd, colStart, colEnd int) {
	if n == nil {
		return 0, 0, 0, 0
	}
	start := fset.Position(n.Pos())
	end := fset.Position(n.End())
	return start.Line, end.Line, start.Column, end.Column
}

func nodeString(fset *token.FileSet, n ast.Node) string {
	if n == nil {
		return ""
	}
	// Keep it simple; AST printing is enough for indexing body snippet in Phase 1.
	var b strings.Builder
	_ = formatNode(&b, fset, n)
	return b.String()
}

func exprString(fset *token.FileSet, e ast.Expr) string {
	if e == nil {
		return ""
	}
	var b strings.Builder
	_ = formatNode(&b, fset, e)
	return b.String()
}

func formatNode(w io.Writer, fset *token.FileSet, n ast.Node) error {
	// Avoid adding new deps; use go/format.
	// Imported via a small local wrapper to keep imports clean.
	return formatNodeImpl(w, fset, n)
}

func extractRefs(info *types.Info, fset *token.FileSet, file *ast.File, filePath string, defUIDByObj map[types.Object]string) ([]RefRecord, error) {
	if info == nil || file == nil || fset == nil {
		return nil, nil
	}
	if defUIDByObj == nil {
		return nil, nil
	}

	var out []RefRecord
	ast.Inspect(file, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if !ok || id == nil {
			return true
		}
		obj := info.Uses[id]
		if obj == nil {
			return true
		}
		toUID, ok := defUIDByObj[obj]
		if !ok || toUID == "" {
			return true
		}
		pos := fset.Position(id.Pos())
		out = append(out, RefRecord{
			ToUID:    toUID,
			FilePath: filepath.Clean(filePath),
			Line:     pos.Line,
			Col:      pos.Column,
		})
		return true
	})
	return out, nil
}
