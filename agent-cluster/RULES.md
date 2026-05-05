# Rules (MUST / MUST NOT)

Các quy tắc này là “nguồn chân lý” để review và triển khai. Nếu task mâu thuẫn với rules, ưu tiên hỏi lại.

## 1) Go Indexer (The Muscle)

### MUST
- Dùng `context.Context` cho mọi I/O (đọc/ghi file, DB, chạy command) hoặc thao tác dài.
- Xử lý lỗi tường minh (`return err`); chỉ dùng `panic` cho tình huống “không thể tiếp tục” ở entrypoint CLI (và phải có message rõ).
- CLI output **JSON ổn định** (schema nhất quán; không đổi field/kiểu dữ liệu tuỳ tiện).
- Tách lớp:
  - `cmd/` chỉ parse flags/args, wiring, exit code.
  - `internal/` chứa core logic (parser/indexer/storage/query) để unit test được.
- Lập chỉ mục phải hỗ trợ **incremental** theo file hashing (không re-index toàn bộ khi không cần).

### MUST NOT
- Không hardcode secrets/paths môi trường đặc thù máy; config qua flags/env.
- Không “refactor diện rộng” ngoài scope task.
- Không thêm dependency mới nếu chưa thống nhất (trừ khi task yêu cầu rõ).

## 2) Python/LangGraph Harness (The Brain)

### MUST
- Tool wrappers chỉ đóng vai trò “adapter” gọi CLI của indexer (không copy logic indexing sang Python).
- State phải tối thiểu gồm:
  - `history`
  - `investigation_path`
  - `current_context`
  - `test_results`

### MUST NOT
- Không chạy lệnh phá hoại repo (reset/clean/rm) nếu không được yêu cầu rõ.

## 3) Testing

### MUST
- Logic nghiệp vụ/core: ưu tiên unit test table-driven.
- Test phải deterministic (không phụ thuộc network).

