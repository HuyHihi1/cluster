# Bài 3: Managed Resources (MR)

**Managed Resource** là đơn vị nhỏ nhất trong Crossplane. Mỗi MR tương ứng 1-1 với một tài nguyên hạ tầng.

## Đặc điểm của MR
- Có cấu hình `forProvider`: Chứa các tham số đặc thù của cloud (ví dụ: size của database).
- Có cấu hình `deletionPolicy`: `Delete` (xóa MR là xóa tài nguyên thật) hoặc `Orphan` (chỉ xóa MR, giữ lại tài nguyên thật).
- Trạng thái `Ready` và `Synced`:
    - `Synced`: Crossplane đã đẩy cấu hình lên cloud thành công.
    - `Ready`: Tài nguyên trên cloud đã sẵn sàng sử dụng.

## Luồng công việc
1. Bạn apply một YAML (ví dụ: `Object` trong S3).
2. Crossplane Provider nhận diện và gọi API của Cloud.
3. Nếu ai đó sửa tài nguyên trên Cloud Console, Crossplane sẽ tự động sửa ngược lại cho đúng với file YAML của bạn.

---
**Tiếp theo:** [Bài 4: Compositions & XRDs](../04-compositions/README.md)
