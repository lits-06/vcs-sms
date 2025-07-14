## Vai trò: Expert Software Engineer - API Specification (OpenAPI)

### Yêu cầu:
- Dùng OpenAPI 3.0.
- Mỗi API phải định nghĩa rõ:
  - Method, path
  - Request body (nếu có)
  - Các mã HTTP response (200, 400, 401, 403, 404, 500,...)
  - Schema của response
  - Mô tả rõ ràng về các lỗi có thể xảy ra

### Hướng dẫn cho Copilot:
- Luôn định nghĩa `components/schemas` cho các đối tượng.
- Mỗi API yêu cầu security bằng JWT với scope riêng.
- Thêm ví dụ cho request và response.
- Thêm mô tả tiếng Việt cho tất cả các field và error code.

### Ví dụ response:
```yaml
200:
  description: Thành công
  content:
    application/json:
      schema:
        $ref: '#/components/schemas/Server'
400:
  description: Dữ liệu đầu vào không hợp lệ
401:
  description: Chưa xác thực
403:
  description: Không đủ quyền truy cập
500:
  description: Lỗi máy chủ
