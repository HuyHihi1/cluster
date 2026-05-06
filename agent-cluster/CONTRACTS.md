# Contracts (CLI + DB)

Tài liệu này là “mỏ neo” để agent không đi chệch hướng khi implement.

## 1) CLI contract (JSON stable)

Các lệnh mục tiêu (theo README):

- `indexer index <path>`: lập chỉ mục codebase.
- `indexer query --name <SymbolName>`: lấy thông tin symbol theo tên.
- `indexer search "<text>"`: full-text search trong code/docstrings.
- (phase nâng cao) `indexer refs --id <symbol_id>`: tìm nơi dùng symbol.

### Nguyên tắc output
- STDOUT: JSON.
- STDERR: logs/hints (nếu cần).
- Exit codes:
  - `0`: success
  - `>0`: error (message rõ, không nuốt lỗi)

### JSON shape (gợi ý chuẩn hoá)

> Lưu ý: đây là “khung” để implement nhất quán; khi bắt đầu code cần chốt chính xác field names/types và **giữ ổn định**.

**`indexer index`**
```json
{
  "ok": true,
  "db_path": ".agent-indexer/index.db",
  "indexed_files": 123,
  "updated_files": 10,
  "skipped_files": 113,
  "duration_ms": 4567
}
```

**`indexer query` / `indexer search`**
```json
{
  "ok": true,
  "results": []
}
```

## 2) DB contract (SQLite + FTS5)

Yêu cầu theo README:
- Lưu symbols + metadata (signature, docstrings, line numbers, file paths).
- Có hashing để incremental indexing.
- Có FTS5 cho full-text (code + docstrings).

### Bảng tối thiểu (định hướng)
- `files`: path, modtime, hash (hoặc content hash), last_indexed_at
- `symbols`: id, name, kind, package, receiver (nếu method), signature, doc, file_path, line_start, line_end, content_hash
- `fts_symbols` (FTS5): index `name`, `signature`, `doc`, và/hoặc snippet code
- (phase nâng cao) `refs`: (symbol_id -> file_path/position) hoặc stored query-able structure

