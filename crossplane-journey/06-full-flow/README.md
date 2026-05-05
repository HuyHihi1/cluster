# Bài 6: Full Flow - Tự xây dựng API riêng của bạn

Trong bài này, chúng ta sẽ thực hiện một quy trình đầy đủ (Full Flow) để tạo ra một dịch vụ trừu tượng.

## Kịch bản
Bạn muốn cấp cho Developer một loại tài nguyên mới tên là `QuickApp`. Khi Developer tạo `QuickApp`, hệ thống sẽ tự động tạo:
1. Một **Namespace** riêng cho app đó.
2. Một **ConfigMap** mẫu bên trong Namespace đó.

## Các bước thực hiện

### Bước 1: Định nghĩa API (XRD)
Chúng ta định nghĩa cấu trúc của `QuickApp`. Nó sẽ nhận vào một tham số là `appName`.
Xem file: `exercises/04-full-flow-definition.yaml` (phần CompositeResourceDefinition).

### Bước 2: Định nghĩa Logic (Composition)
Chúng ta dạy Crossplane rằng: "Nếu thấy ai đó gọi QuickApp, hãy dùng Provider Kubernetes để tạo ra 1 Namespace và 1 ConfigMap".
Xem file: `exercises/04-full-flow-definition.yaml` (phần Composition).

### Bước 3: Sử dụng (Claim)
Developer chỉ cần tạo một file đơn giản để yêu cầu dịch vụ.
Xem file: `exercises/05-full-flow-claim.yaml`.

---
## Cách chạy
Tôi đã thêm task vào Taskfile để bạn thực hiện nhanh:
1. `task 15-full-flow-setup`: Apply XRD và Composition.
2. `task 16-full-flow-claim`: Tạo Claim (yêu cầu tài nguyên).
3. `task 17-full-flow-check`: Kiểm tra xem Namespace và ConfigMap đã được tạo tự động chưa.
