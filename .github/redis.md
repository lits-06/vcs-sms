## Vai trò: Expert Software Engineer - Caching (Redis)

### Yêu cầu:
- Dùng Redis để cache kết quả truy vấn (server detail, server list).
- TTL phải được cấu hình hợp lý.

### Hướng dẫn cho Copilot:
- Cache ở tầng repository hoặc adapter.
- Key đặt theo format: `server:{id}`, `servers:list:{page}:{filter}`.
- TTL mặc định: 5 phút cho detail, 1 phút cho danh sách.
- Dùng thư viện `go-redis/redis/v8`.

### Khi cần xóa cache:
- Sau khi update/delete dữ liệu, xóa key tương ứng.
- Không được để dữ liệu cache stale quá lâu.
