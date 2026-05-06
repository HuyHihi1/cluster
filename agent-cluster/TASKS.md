# Backlog (Atomic Tasks)

Chỉ làm **1 task/lần**. Mỗi task phải có output kiểm được (build/test/CLI).

## Phase 1: Foundation (Go Indexer Core)

1) Thiết kế SQLite schema (symbols, FTS5, hashing)
- Output: `docs/schema.sql` + giải thích ngắn về bảng/indices

2) Triển khai parser dùng `golang.org/x/tools/go/packages`
- Output: extract được symbols + metadata tối thiểu

3) Lập chỉ mục full-text cho docstrings và mã nguồn
- Output: query FTS trả về kết quả đúng

4) CLI cơ bản: `indexer index <path>`
- Output: JSON summary + tạo DB ở `.agent-indexer/index.db`

## Phase 2: Advanced Indexing & API

5) Incremental indexing theo hashing (chỉ update file thay đổi)

6) Query interface trả JSON: `query/search`

7) (Nâng cao) Cross-reference: find usages / references

## Phase 3: AI Harness (LangGraph)

8) Định nghĩa `AgentState`

9) Python tool wrappers gọi Go indexer CLI
- Tools: `search_symbols`, `get_implementation`, `find_references`, `run_tests`

10) Triển khai các node + conditional edges (đủ thông tin thì dừng tìm kiếm)

## Phase 4: Closing the Loop (Verify & Fix)

11) Tool `run_tests(package)` + node verification

12) Tích hợp cơ chế apply code changes (giới hạn scope) + tối ưu prompt/graph

