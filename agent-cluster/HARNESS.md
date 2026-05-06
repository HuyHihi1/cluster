# Harness (Prompt Template) cho Go-Agent-Indexer

Mục tiêu: giảm lặp prompt, tăng tính “contract-first”, và giúp agent làm việc theo từng **atomic task**.

## Cách dùng nhanh

Khi giao việc, chỉ cần gửi:

- “Dựa theo `RULES.md` + `CONTRACTS.md`, hãy làm `TASKS.md` task #N.”
- Kèm input đặc thù (log, file path, yêu cầu mới) nếu có.

## Prompt template (copy/paste)

```text
ROLE
- Bạn là Senior Engineer (Go + Agent Systems). Ưu tiên: đúng contract, test được, incremental indexing.

PROJECT SNAPSHOT
- Repo: agent-cluster
- Stack (local env):
  - Go 1.24.3
  - Python 3.13.3
  - SQLite (FTS5 enabled)
  - LangGraph: latest compatible (pin khi cài)
- Architecture:
  - Layer 1 (Go Indexer): parse + index symbols -> SQLite/FTS5; CLI output JSON
  - Layer 2 (Python Harness): LangGraph orchestration; gọi indexer qua CLI wrappers
- Default DB path: .agent-indexer/index.db (trong repo)
- Non-goals: không refactor ngoài scope task; không thêm dependency mới nếu chưa được chốt.

ANCHORS (không được phá)
- Rules: đọc và tuân thủ RULES.md
- Contracts: đọc và tuân thủ CONTRACTS.md (CLI JSON + DB schema)
- Backlog: thực hiện theo TASKS.md (1 task/lần)

TASK
- Thực hiện: TASKS.md task #[N]
- Output:
  - Patch code (chỉ trong repo)
  - Lệnh verify (build/test) tương ứng

WORKFLOW
- Nếu thiếu info “blocking”: hỏi tối đa 3 câu rồi dừng.
- Nếu không thiếu: làm luôn, báo kết quả + cách verify.
```

