# Bài 4: Compositions & XRDs

Đây là tính năng mạnh mẽ nhất của Crossplane, giúp hiện thực hóa tư tưởng **Platform Engineering**.

## Thử thách
Giả sử công ty bạn có 100 team. Nếu team nào cũng phải tự viết YAML cho RDS, Subnet, IAM Role... thì sẽ rất lộn xộn và dễ sai sót.

## Giải pháp: Composition
Bạn (Platform Admin) sẽ tạo ra một "gói Combo":
1. **XRD (Composite Resource Definition)**: Định nghĩa "Cái vỏ". Ví dụ: Tôi muốn một loại tài nguyên tên là `MyPostgres`, chỉ cần truyền vào `diskSize`.
2. **Composition**: Định nghĩa "Ruột". Nó sẽ nhận `diskSize` từ `MyPostgres` và mapping vào các tài nguyên thực tế (RDS, Monitoring, CloudWatch...).

## Lợi ích
- **Abstraction**: Che giấu sự phức tạp.
- **Enforcement**: Ép các tiêu chuẩn bảo mật (ví dụ: luôn phải encrypt database).
- **Multi-cloud**: Bạn có thể tạo 2 Compositions cho cùng 1 XRD (Một cái chạy AWS, một cái chạy Azure).

---
**Tiếp theo:** [Bài 5: Claims](../05-claims/README.md)
