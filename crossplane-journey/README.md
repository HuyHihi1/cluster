# Lộ trình học Crossplane 🚀

Crossplane là một công cụ mã nguồn mở giúp biến Kubernetes cluster của bạn thành một **Universal Control Plane**. Thay vì chỉ quản lý container, bạn có thể quản lý cả database, storage, network từ các cloud provider (AWS, GCP, Azure) hoặc bất kỳ service nào có API.

## Mục đích của Crossplane
1. **Universal Control Plane**: Quản lý mọi loại hạ tầng (S3, RDS, IAM...) bằng chính các câu lệnh `kubectl` quen thuộc.
2. **Platform Engineering & Abstraction**: Cho phép đội Platform xây dựng các API nội bộ đơn giản cho Developer sử dụng (Self-service), che giấu sự phức tạp của Cloud.
3. **Continuous Reconciliation**: Tự động phát hiện và sửa lỗi (Drift Detection) 24/7 để đảm bảo hạ tầng luôn đúng với thiết kế ban đầu.

## Mục lục bài học

1.  **[Bài 1: Giới thiệu Crossplane](./01-introduction/README.md)**
    *   Control Plane là gì?
    *   Tại sao chọn Crossplane thay vì Terraform?
    *   Các khái niệm cốt lõi: Providers, Managed Resources, Compositions.
2.  **[Bài 2: Cài đặt và Cấu hình](./02-setup/README.md)**
    *   Cài đặt Crossplane lên K8s (Kind/Minikube).
    *   Cài đặt Provider đầu tiên.
3.  **[Bài 3: Managed Resources (MR)](./03-managed-resources/README.md)**
    *   Quản lý tài nguyên hạ tầng đơn lẻ qua YAML.
    *   Cơ chế Reconcile và Drift Detection.
4.  **[Bài 4: Compositions & XRDs](./04-compositions/README.md)**
    *   Xây dựng nền tảng (Platform Engineering).
    *   Đóng gói nhiều tài nguyên phức tạp thành một API đơn giản.
5.  **[Bài 5: Claims](./05-claims/README.md)**
    *   Self-service cho Developer.
    *   Cách request tài nguyên từ Platform API.

## Bài tập thực hành
Tất cả bài tập nằm trong thư mục `exercises/`.

---
*Chúc bạn học tốt! Nếu có thắc mắc gì cứ hỏi mình nhé.*
