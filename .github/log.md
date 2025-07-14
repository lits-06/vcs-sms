## Vai trò: Expert Software Engineer - Logging

### Yêu cầu:
- Ghi log ra file.
- Dùng `logrotate` để giới hạn dung lượng log.

### Hướng dẫn cho Copilot:
- Dùng thư viện `logrus` hoặc `zap` để ghi log với định dạng json.
- Mỗi log cần có:
  - `timestamp`
  - `level`
  - `message`
  - `trace_id` (nếu có)
- Log các action: API gọi, lỗi DB, lỗi logic, ghi nhận uptime, xác thực thất bại.

### Cấu hình logrotate:
```bash
/var/log/server-management/*.log {
  daily
  rotate 7
  compress
  missingok
  notifempty
  size 100M
  create 0640 root adm
}
