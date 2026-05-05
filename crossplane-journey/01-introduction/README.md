# Bài 1: Crossplane là gì?

## 1. Khái niệm Control Plane
Trong thế giới Kubernetes, **Control Plane** là bộ não quản lý trạng thái của cluster. Crossplane mở rộng khả năng này để quản lý **mọi thứ** bên ngoài cluster.

## 2. Crossplane vs Terraform
| Đặc điểm | Terraform | Crossplane |
| :--- | :--- | :--- |
| **Mô hình** | Infrastructure as Code (Push) | Infrastructure as Data (Pull/Control Plane) |
| **State** | Lưu ở file `.tfstate` | Lưu trực tiếp trong Kubernetes etcd |
| **Drift Detection** | Phải chạy `plan` mới thấy | Tự động phát hiện và sửa lỗi liên tục (Reconcile) |
| **Self-service** | Khó (cần CI/CD phức tạp) | Dễ (thông qua Kubernetes API/Claims) |

## 3. Các thành phần chính
- **Provider**: Các "plugin" để Crossplane nói chuyện với cloud (AWS, GCP, Azure, Helm, SQL...).
- **Managed Resource (MR)**: Đại diện cho 1 tài nguyên thực tế (ví dụ: một cái S3 Bucket).
- **Composite Resource Definition (XRD)**: Định nghĩa "API ảo" của bạn (ví dụ: `XDatabase`).
- **Composition**: Cách hiện thực hóa `XRD` (ví dụ: `XDatabase` = `RDS Instance` + `DB Subnet Group` + `Security Group`).
- **Claim**: Yêu cầu từ phía người dùng cuối (Developer) để tạo tài nguyên.

---
**Tiếp theo:** [Bài 2: Cài đặt và Cấu hình](../02-setup/README.md)
