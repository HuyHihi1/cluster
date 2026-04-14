# 📚 Tổng hợp Kiến thức: Go Observability Playground

File này dùng để lưu trữ lại tất cả những kiến thức, khái niệm và các kinh nghiệm sửa lỗi (troubleshooting) quý giá đã được rút ra trong suốt quá trình xây dựng dự án Playground này.

---

## 1. Các Khái niệm Cốt lõi (Core Concepts)

### A. Structured Logging (Log có cấu trúc)
Thay vì sử dụng các dòng log dạng text thô (`fmt.Println`), chúng ta ưu tiên sử dụng `log/slog` (đã được tích hợp sẵn từ Go 1.21). 
- Giúp xuất log dưới định dạng **JSON** (`slog.NewJSONHandler`).
- Hỗ trợ truyền `context.Context` (để đính kèm TraceID, SpanID vào log, giúp dễ dàng map Log với Trace).

### B. Custom Metrics (Đo lường với Prometheus)
Sử dụng thư viện `prometheus/client_golang` để tạo ra các bộ đếm tùy chỉnh:
- **Counter**: Đếm tổng số lượng (ví dụ: Tổng số HTTP request).
- **Histogram/Summary**: Theo dõi độ trễ hoặc phân phối dữ liệu (thời gian phản hồi nhanh hay chậm).
- Dữ liệu được public thông qua endpoint `/metrics` và được Prometheus tới kéo (scrape) định kỳ.

### C. Distributed Tracing (Dấu vết phân tán với OpenTelemetry & Tempo)
- **Truy vết**: Thay vì chỉ biết API báo lỗi, Tracing giúp ta thấy được "Biểu đồ thác nước" (Waterfall đồ thị) chỉ ra thời gian tiêu tốn ở từng bước nhỏ (ví dụ: mất 5ms cho mã hóa, 150ms cho câu query Database).
- **Tiêu chuẩn (Standard)**: Chúng ta sử dụng chuẩn **OpenTelemetry (OTel)** thay vì các chuẩn cũ như OpenTracing hay Jaeger Client. Dữ liệu được gửi qua giao thức OTLP (gRPC hoặc HTTP) tới collector (ở đây là **Grafana Tempo**).

### D. Profiling (Lấy mẫu luồng thực thi với pprof)
`pprof` là "bảo bối" của Go để tối ưu hệ thống:
- Cắm vào API qua `net/http/pprof`.
- **CPU Profile**: Tìm ra chính xác dòng code/hàm nào đang ngốn phần trăm CPU cao nhất thông qua Flame Graph.
- **Heap Profile**: Tìm ra điểm gây rò rỉ bộ nhớ (Memory Leak) hoặc hàm nào đang cấp phát quá nhiều RAM vô nghĩa.

### E. Continuous Profiling (Grafana Pyroscope)
Thay vì gõ lệnh lưu file thủ công mỗi khi Server chậm, **Pyroscope** hoạt động như một con bot tự động âm thầm truy xuất endpoint pprof mỗi 10 giây và lưu cục bộ. Giúp lập trình viên có thể mở Grafana và xem lịch sử nghẽn cổ chai của Server ở bất kỳ thời điểm nào trong quá khứ.

### F. Benchmarking
Sử dụng chuẩn `testing.B` của Go để đo đạc vi mô sự tối ưu thuật toán.
Chúng ta có thể đo độ tốn CPU và độ cấp phát RAM cùng lúc thông qua tham số `-benchmem`. Đặc biệt, có thể sinh ra file profile offline bằng tham số `-cpuprofile`.

---

## 2. Best Practices (Thực hành Tốt Nhất cho OTel)
Thông qua phân tích cấu trúc mã nguồn OTel chuyên nghiệp, chúng ta đã đúc kết được các Best Practices sau khi triển khai hệ thống cho Production:
1. **Dùng biến môi trường (Environment Variables)**: Thay vì gán cứng cấu hình ServiceName vào code, hãy gọi `resource.WithFromEnv()`. Bộ OTel SDK sẽ tự động đọc các biến môi trường cấu hình như `OTEL_RESOURCE_ATTRIBUTES` để nhúng vào hệ thống.
2. **Graceful Shutdown**: Gom tất cả các hàm dọn dẹp (`TracerProvider.Shutdown`, `MeterProvider.Shutdown`) vào một hàm `errors.Join` duy nhất và gọi vào đúng lúc Server tắt, đảm bảo không có telemetry nào bị thất thoát (dropped) chưa kịp gửi.
3. **Internal Logger**: Liên kết bộ log hệ thống bắt lỗi (`logr`) vào `otel.SetLogger` để ghi nhận lại nếu client của OTel không kết nối được tới máy chủ (OTLP network drops).

---

## 3. Nhật ký Sửa lỗi (Troubleshooting & Bugs)

Trong quá trình xây dựng, chúng ta đã giải quyết được một số bài toán hóc búa liên quan đến cấu hình hệ thống:

### Bug #1: Tempo Configurations (Phiên bản mới)
- **Hiện tượng**: Docker container `tempo` bị CrashLoop ngay khi khởi động.
- **Nguyên nhân**: Grafana Tempo thay đổi liên tục syntax của file `tempo.yaml` qua các phiên bản. Các key phụ trợ như `ingester` hay `compactor` nếu thiết lập sai sẽ khiến service từ chối khởi động. Lỗi văng báo thiếu Kafka topic cũng xuất phát từ việc cấu hình distributor/backend không tương thích.
- **Khắc phục**: Chúng ta đã loại bỏ các trường lỗi, đồng thời **Ghim (Pin) cứng version image là `grafana/tempo:2.3.0`** vào file Docker Compose. Điều này giúp hệ thống vĩnh viễn không bị lỗi vặt khi image latest bị Grafana cập nhật thay đổi đột ngột.

### Bug #2: Grafana Pyroscope Datasource (Unsupported Protocol Scheme "")
- **Hiện tượng**: API Grafana liên tục trả về HTTP 500 kèm thông báo lỗi `"unsupported protocol scheme \"\""` khi bấm chọn Pyroscope.
- **Nguyên nhân cốt lõi**:
  1. Loại Datasource (Datasource Type) của Pyroscope thay đổi. Trước đây nó tên là `phlare` hoặc loại ứng dụng ngoài là `grafana-pyroscope-datasource`. Ở bản mới, đôi khi lỗi định dạng type này dắt tới thiết lập sai.
  2. Việc sửa đổi file `grafana-datasources.yml` và khởi động lại container grafana tạo ra cấu hình Datasource mới (Healthy, ReadOnly), tuy nhiên Grafana DBMS (SQLite) vẫn không xóa bỏ record datasource cũ bị lỗi (bị rỗng URL) từ lần cấu hình sai lầm trước đó. ID Datasource cũ kẹt trong bộ nhớ Cache / URL của Client-side.
- **Khắc phục**: Sử dụng **Grafana HTTP API** để tiến hành `DELETE` chính xác `uid` của Datasource bóng ma kia. Khôi phục lại đúng `type: grafana-pyroscope-datasource` và tải lại giao diện hoàn toàn (Hard Reload). Mọi dấu vết lỗi đều bị xóa sạch!

### Bug #3: Broken Trace Propagation across NATS
- **Hiện tượng**: Xuất hiện các Trace ID riêng lẻ cho Server A và Server B thay vì một trace duy nhất nối liền từ HTTP -> NATS -> Worker.
- **Nguyên nhân & Khắc phục**:
  1. **Headers Initialization**: Trong Go, `msg.Header` của NATS mặc định là `nil`. Việc gọi `Inject` vào một map nil sẽ không có tác dụng. Cần khởi tạo bằng `msg.Header = make(nats.Header)` trước khi inject.
  2. **Type Compatibility**: `propagation.HeaderCarrier` chỉ dành riêng cho `http.Header`. Mặc dù `nats.Header` có cấu trúc tương tự (`map[string][]string`), chúng là hai kiểu dữ liệu khác nhau trong Go.
  - **Giải pháp**: Triển khai một **Shim (NatsCarrier)** thực thi interface `propagation.TextMapCarrier` để OpenTelemetry có thể đọc/ghi chính xác vào header của NATS.

---

## 4. In-Band Propagation (Truyền tin nội bộ)

Khi làm việc với các Message Broker, việc truyền Trace Context (TraceID, SpanID) được gọi là **In-Band Propagation**.

- **Cơ chế**: Sử dụng các trường "Metadata" đi kèm message thay vì nhét vào trong body (payload) của dữ liệu business.
- **Kafka**: Sử dụng **Record Headers** (từ v0.11).
- **Pulsar**: Sử dụng **Properties** (một map string-string).
- **Lợi ích**: Consumer có thể biết được TraceID mà không cần giải mã (deserialization) toàn bộ nội dung message, giúp hệ thống tracing hoạt động độc lập và hiệu quả.
- **Lưu ý**: Luôn sử dụng hàm `Inject` và `Extract` của OTel thay vì gán thủ công để đảm bảo tuân thủ tiêu chuẩn W3C Trace Context (bao gồm cả các cờ như sampling).
