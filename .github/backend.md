## Vai trò: Expert Software Engineer - Backend

### Công nghệ:
- Ngôn ngữ: Golang
- Framework: Gin
- Tuân thủ nguyên lý Clean Architecture:
  - Tách biệt rõ `domain`, `usecase`, `interface`, `infrastructure`.
  - Không phụ thuộc chiều ngược (chỉ interface phụ thuộc vào logic).

### Hướng dẫn cho Copilot:
- Tạo route và handler API trong layer `interface`.
- Xử lý logic nghiệp vụ trong `usecase`.
- Truy cập dữ liệu thông qua interface được inject vào usecase.
- Middleware phải xử lý xác thực JWT và kiểm tra scope cụ thể cho từng API.
- Sử dụng DI (Dependency Injection) để truyền các service/infrastructure cần thiết.

### Ví dụ cụ thể:
- Khi viết API `GET /servers/:id`:
  - Middleware kiểm tra token và scope: `server:read`.
  - Handler gọi usecase `GetServerByID`.
  - Usecase gọi repository interface `ServerRepository`.
  - Repository thực thi query ở tầng infrastructure.
