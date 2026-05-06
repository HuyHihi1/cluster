# Go-Agent-Indexer: local Code Intelligence for Autonomous Agents

Dự án này là một giải pháp toàn diện cho việc xây dựng một AI Agent chuyên biệt cho ngôn ngữ Go, có khả năng tự động điều tra bug và thực hiện tính năng bằng cách kết hợp sức mạnh của một **High-Performance Indexer** và một **Autonomous Agent Harness**.

---

## 🏗️ Kiến trúc tổng thể (High-Level Architecture)

Hệ thống được chia thành hai lớp (Layers) tách biệt nhưng phối hợp chặt chẽ:

1.  **Indexer Layer (Golang)**: Chịu trách nhiệm "thấu thị" codebase. Nó quét, phân tích AST và lưu trữ mọi Symbol vào SQLite.
2.  **Intelligence Layer (Python/LangGraph)**: Chịu trách nhiệm "suy luận". Nó sử dụng Indexer làm công cụ để tìm kiếm ngữ cảnh và đưa ra giải pháp.

---

## 🛠️ Thành phần 1: Go Code Indexer (The Muscle)

Bộ chỉ mục hiệu năng cao, chạy local 100%.

### Tính năng chính:

- **Dependency Discovery**: Sử dụng `golang.org/x/tools/go/packages` để nhận diện workspace, Go modules phụ thuộc (`GOMODCACHE`) và thư viện chuẩn (stdlib) với đầy đủ thông tin về Type và Imports.
- **Semantic Parsing**: Phân tích sâu:
  - Symbols: Functions, Structs, Interfaces, Methods, Types.
  - Metadata: Signatures, Docstrings, Line numbers, File paths, Hashing (cho Incremental Indexing).
- **Incremental Indexing**: Chỉ quét và cập nhật những file có thay đổi dựa trên file hashing, giúp tiết kiệm tài nguyên.
- **Fast Storage**: Sử dụng **SQLite với FTS5** để lập chỉ mục toàn văn (Full-text search) cho cả code và tài liệu (docstrings).
- **Output**: JSON-structured output qua CLI, giúp dễ dàng tích hợp với các ngôn ngữ khác (Python/Nodejs).

### Các lệnh CLI mẫu:

- `indexer index ./...`: Bắt đầu lập chỉ mục toàn bộ dự án.
- `indexer query --name ProcessOrder`: Tìm thông tin về hàm/struct.
- `indexer search "nil pointer error"`: Tìm kiếm Full-text trong code và tài liệu.

---

## 🧠 Thành phần 2: AI Agent Harness (The Brain)

Sử dụng **LangGraph** để xây dựng bộ não điều phối cho Agent.

### 1. Core Tools (Công cụ của Agent)

Agent được trang bị các công cụ để tương tác với Indexer và Hệ thống:

- `search_symbols(query)`: Tìm kiếm tên hàm/struct từ SQLite.
- `get_implementation(symbol_id)`: Lấy mã nguồn chi tiết của một symbol cụ thể.
- `find_references(symbol_id)`: Tìm tất cả các nơi đang sử dụng symbol này.
- `run_tests(package)`: Chạy `go test` để xác thực lỗi hoặc kiểm tra bản fix.

### 2. State Management (Trạng thái Agent)

Agent lưu trữ một `State` bao gồm:

- `history`: Lịch sử hội thoại và suy luận.
- `investigation_path`: Danh sách các file/hàm đã kiểm tra.
- `current_context`: Các đoạn code (snippets) liên quan nhất.
- `test_results`: Kết quả chạy thử nghiệm để quyết định bước tiếp theo.

### 3. Thiết kế Đồ thị suy luận (Reasoning Graph)

Agent chạy theo chu kỳ suy luận, thử nghiệm và sửa lỗi:

```mermaid
graph TD
    START((Bắt đầu)) --> ReceiveBug[Tiếp nhận Log/Yêu cầu]
    ReceiveBug --> AnalyzeLog[Node: Phân tích Log]
    AnalyzeLog --> SearchIndexer[Node: Truy vấn Indexer]
    SearchIndexer --> CheckCode[Node: Đọc & Hiểu Code]
    CheckCode --> Decision{Đủ thông tin?}
    Decision -- Chưa --> SearchIndexer
    Decision -- Rồi --> ProposeFix[Node: Đề xuất Fix/Feature]
    ProposeFix --> RunVerify[Node: Chạy Test/Verify]
    RunVerify --> Success{Pass Test?}
    Success -- No --> AnalyzeLog
    Success -- Yes --> END((Kết thúc))
---

## 🚀 Lộ trình triển khai (Implementation Roadmap)

Dự án được chia thành 4 giai đoạn chính:

### Phase 1: Foundation (Go Indexer Core)
- [ ] Thiết kế Schema SQLite (Symbols, FTS5, Hashing).
- [ ] Triển khai Parser sử dụng `golang.org/x/tools/go/packages`.
- [ ] Lập chỉ mục Full-text search cho Docstrings và Mã nguồn.
- [ ] CLI cơ bản: `indexer index <path>`.

### Phase 2: Advanced Indexing & API
- [ ] Cơ chế Incremental Indexing (chỉ quét file có thay đổi).
- [ ] Query Interface: Trả về kết quả JSON cho Agent (Search, Get, List).
- [ ] (Nâng cao) Cross-reference: Tìm kiếm các nơi sử dụng Symbol (Find Usages).

### Phase 3: AI Harness (LangGraph)
- [ ] Định nghĩa `AgentState` và cấu trúc đồ thị suy luận.
- [ ] Viết Python Tool Wrappers để giao tiếp với Go Indexer qua CLI.
- [ ] Triển khai các Node: Phân tích Log, Truy vấn, Đọc code.
- [ ] Logic điều kiện (Conditional Edges) để Agent tự quyết định khi nào dừng tìm kiếm.

### Phase 4: Closing the Loop (Verify & Fix)
- [ ] Tool: `run_tests` để tự động chạy bộ kiểm thử của Go.
- [ ] Node: Verification để xác thực bản fix dựa trên kết quả Test.
- [ ] Tích hợp cơ chế áp dụng mã nguồn (Code Writing) vào file.
- [ ] Tối ưu Prompt và luồng suy luận để giảm thiểu chi phí token.

```

---

## 📎 Harness files (để giảm lặp prompt)

- `HARNESS.md`: template prompt để giao task.
- `RULES.md`: bộ quy tắc MUST/MUST NOT (chuẩn review/implement).
- `CONTRACTS.md`: mỏ neo contract (CLI JSON + DB định hướng).
- `TASKS.md`: backlog atomic tasks theo roadmap.
