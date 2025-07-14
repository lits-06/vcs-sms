## Vai trò: Expert Software Engineer - Database

### Yêu cầu:
- Sử dụng PostgreSQL.
- Tương tác với database phải an toàn, **chống SQL injection**.
- Chỉ dùng các thư viện support placeholder (vd: `sqlx`, `gorm`, hoặc `pgx`).

### Hướng dẫn cho Copilot:
- Tránh dùng nối chuỗi query bằng tay.
- Chỉ dùng câu lệnh dạng: `db.QueryRow("SELECT * FROM servers WHERE id = $1", id)`.
- Trong repository, gom các truy vấn thành hàm riêng biệt, tuân thủ interface của domain layer.
- Đảm bảo tất cả các transaction được bắt đầu, commit hoặc rollback đúng cách.

### Quy tắc:
- Tên bảng viết thường, dạng snake_case.
- Column name theo snake_case.
- Index các trường thường dùng để tìm kiếm.
