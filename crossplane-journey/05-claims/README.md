# Bài 5: Claims (Yêu cầu tài nguyên)

**Claim** là cách mà Developer tương tác với Platform.

## Tại sao cần Claim?
Trong K8s, thường thì Platform Admin quản lý các tài nguyên mang tính global (Cluster-scoped). Developer chỉ có quyền trong `namespace` của họ.

- **XRD/Composition**: Thường là Cluster-scoped (Admin quản lý).
- **Claim**: Là Namespace-scoped. Developer tạo Claim trong namespace của họ, Crossplane sẽ tự động sinh ra các tài nguyên tương ứng ở tầng Cluster.

## Ví dụ
Developer muốn database, họ chỉ cần tạo một file đơn giản:
```yaml
apiVersion: database.example.org/v1alpha1
kind: PostgreSQLInstance
metadata:
  namespace: app-team-a
  name: my-db
spec:
  parameters:
    storageGB: 20
```
Crossplane sẽ lo phần còn lại.

---
**Chúc mừng!** Bạn đã nắm được luồng đi cơ bản của Crossplane. Hãy bắt đầu làm bài tập tại thư mục `exercises/` nhé.
