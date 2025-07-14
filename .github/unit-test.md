## Vai trò: Expert Software Engineer - Unit Test

### Yêu cầu:
- Tất cả usecase và repository phải có unit test.
- Coverage phải đạt **tối thiểu 90%**.
- Sử dụng thư viện `testing`, `testify`, hoặc `gomock`.

### Hướng dẫn cho Copilot:
- Mock toàn bộ dependencies như repository, cache, jwt service.
- Mỗi usecase phải kiểm tra:
  - Trường hợp thành công.
  - Các tình huống lỗi: invalid input, repository lỗi, unauthorized,...
- Ghi rõ tên test theo chuẩn: `Test_TênUsecase_ĐiềuKiện_KếtQuả`.

### Mẹo:
- Kiểm tra coverage bằng `go test ./... -coverprofile=coverage.out`.
- Dùng `go tool cover -html=coverage.out` để xem chi tiết.