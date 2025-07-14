## Vai trò: Expert Software Engineer - Elasticsearch

### Yêu cầu:
- Dùng Elasticsearch để lưu log uptime server theo thời gian.
- Hỗ trợ truy vấn theo time range.

### Hướng dẫn cho Copilot:
- Mỗi lần ghi uptime, gửi log gồm: `server_id`, `status`, `timestamp`, `latency`.
- Sử dụng index: `uptime-logs-YYYY.MM.DD`.
- Mapping trường `timestamp` là `date`, `latency` là `float`.

### Truy vấn mẫu:
- Tìm % thời gian online trong 1 ngày => đếm số `status == "up"` chia cho tổng log.
- Tìm thời điểm downtime lâu nhất của server theo `server_id`.
