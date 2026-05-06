-- SQLite schema for agent-cluster indexer (Phase 1: Foundation)
-- Targets: incremental indexing (by file hashing) + symbol metadata + FTS5 search.
--
-- Notes:
-- - The application is responsible for opening the DB with `PRAGMA foreign_keys=ON;`.
-- - This file is intended to be applied to an empty SQLite database.

PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA foreign_keys = ON;

-- Tracks schema compatibility for future migrations.
CREATE TABLE IF NOT EXISTS meta (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
);

INSERT INTO meta(key, value) VALUES ('schema_version', '2')
ON CONFLICT(key) DO UPDATE SET value=excluded.value;

-- One row per indexed file. Used for incremental indexing.
CREATE TABLE IF NOT EXISTS files (
  path            TEXT PRIMARY KEY, -- repo-relative or absolute (decide in app config, but keep consistent)
  size_bytes      INTEGER NOT NULL,
  mtime_unix_ms   INTEGER NOT NULL,  -- last observed modtime from filesystem
  content_hash    TEXT NOT NULL,     -- hash of file content (e.g., sha256 hex)
  language        TEXT NOT NULL DEFAULT '', -- optional hint: go/py/md/...
  last_indexed_at INTEGER NOT NULL    -- unix ms when indexer last processed this file
);

CREATE INDEX IF NOT EXISTS idx_files_content_hash ON files(content_hash);
CREATE INDEX IF NOT EXISTS idx_files_last_indexed_at ON files(last_indexed_at);

-- One row per symbol discovered from source files.
CREATE TABLE IF NOT EXISTS symbols (
  id            INTEGER PRIMARY KEY,
  uid           TEXT NOT NULL UNIQUE, -- stable identifier (e.g., hash of (file_path + name + kind + span))
  file_path     TEXT NOT NULL,
  name          TEXT NOT NULL,
  kind          TEXT NOT NULL,        -- func/type/method/var/const/field/package/...
  package       TEXT NOT NULL DEFAULT '',
  receiver      TEXT NOT NULL DEFAULT '', -- for methods
  signature     TEXT NOT NULL DEFAULT '',
  doc           TEXT NOT NULL DEFAULT '',
  body          TEXT NOT NULL DEFAULT '', -- optional: snippet/source for search
  line_start    INTEGER NOT NULL,
  line_end      INTEGER NOT NULL,
  col_start     INTEGER NOT NULL DEFAULT 0,
  col_end       INTEGER NOT NULL DEFAULT 0,
  content_hash  TEXT NOT NULL DEFAULT '', -- hash of symbol-relevant content for incremental symbol updates
  updated_at    INTEGER NOT NULL,         -- unix ms
  FOREIGN KEY (file_path) REFERENCES files(path) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_symbols_file_path ON symbols(file_path);
CREATE INDEX IF NOT EXISTS idx_symbols_name ON symbols(name);
CREATE INDEX IF NOT EXISTS idx_symbols_kind_name ON symbols(kind, name);
CREATE INDEX IF NOT EXISTS idx_symbols_content_hash ON symbols(content_hash);

-- Full-text index for symbols. External-content mode keeps the canonical data in `symbols`.
-- `rowid` == symbols.id for fast joins.
CREATE VIRTUAL TABLE IF NOT EXISTS fts_symbols USING fts5(
  name,
  signature,
  doc,
  body,
  tokenize = 'unicode61',
  content = 'symbols',
  content_rowid = 'id'
);

-- Sync triggers for `fts_symbols`.
CREATE TRIGGER IF NOT EXISTS trg_symbols_ai AFTER INSERT ON symbols BEGIN
  INSERT INTO fts_symbols(rowid, name, signature, doc, body)
  VALUES (new.id, new.name, new.signature, new.doc, new.body);
END;

CREATE TRIGGER IF NOT EXISTS trg_symbols_ad AFTER DELETE ON symbols BEGIN
  INSERT INTO fts_symbols(fts_symbols, rowid, name, signature, doc, body)
  VALUES ('delete', old.id, old.name, old.signature, old.doc, old.body);
END;

CREATE TRIGGER IF NOT EXISTS trg_symbols_au AFTER UPDATE ON symbols BEGIN
  INSERT INTO fts_symbols(fts_symbols, rowid, name, signature, doc, body)
  VALUES ('delete', old.id, old.name, old.signature, old.doc, old.body);
  INSERT INTO fts_symbols(rowid, name, signature, doc, body)
  VALUES (new.id, new.name, new.signature, new.doc, new.body);
END;

-- Cross references (usages). Populated in Phase 2+.
CREATE TABLE IF NOT EXISTS refs (
  id         INTEGER PRIMARY KEY,
  symbol_id  INTEGER NOT NULL,
  file_path  TEXT NOT NULL,
  line       INTEGER NOT NULL,
  col        INTEGER NOT NULL,
  FOREIGN KEY (symbol_id) REFERENCES symbols(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_refs_symbol_id ON refs(symbol_id);
CREATE INDEX IF NOT EXISTS idx_refs_file_path ON refs(file_path);
