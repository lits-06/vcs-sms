
@workspace api tính uptime yêu cầu input là thời gian bắt đầu, thời gian kết thúc, output là gửi email kết quả đến admin, nên chia thành 2 service uptime và mail hay gộp thành 1 service uptime

> Copilot: Dựa trên codebase hiện tại và yêu cầu API tính uptime, tôi khuyến nghị **chia thành 2 service riêng biệt**: `UptimeService` và `MailService`. Đây là cách tiếp cận tốt hơn vì:
> 
> ## Ưu điểm của việc tách riêng:
> 
> ### 1. **Single Responsibility Principle**
> - `UptimeService` chỉ tập trung vào việc tính toán thống kê uptime
> - `MailService` chỉ tập trung vào việc gửi email
> 
> ### 2. **Tái sử dụng cao**
> - `UptimeService` có thể được sử dụng cho các API khác (dashboard, reporting)
> - `MailService` có thể gửi email cho nhiều mục đích khác (alert, notification)
> 
> ### 3. **Dễ test và maintain**
> - Mock từng service độc lập
> - Thay đổi logic email không ảnh hưởng đến tính toán uptime
> 
> ## Cấu trúc đề xuất:
> 
> ### 1. Cập nhật interface.go:
> 
> ```go
> // Thêm interface mới
> type MailService interface {
>     SendUptimeReport(ctx context.Context, stats *UptimeStats, email string) error
> }
> 
> // Cập nhật UptimeService interface
> type UptimeService interface {
>     GetUptimeStats(ctx context.Context, req *UptimeRequest) (*UptimeStats, error)
>     SendUptimeReportByEmail(ctx context.Context, req *UptimeRequest) error
> }
> ```
> 
> ### 2. Tạo mail.go:
> 
> ```go
> package server
> 
> import (
>     "context"
>     "fmt"
>     "github.com/lits-06/vcs-sms/config"
>     "github.com/lits-06/vcs-sms/pkg/logger"
>     "gopkg.in/gomail.v2"
> )
> 
> type mailService struct {
>     config *config.SMTPConfig
>     logger logger.Logger
> }
> 
> func NewMailService(config *config.SMTPConfig, logger logger.Logger) MailService {
>     return &mailService{
>         config: config,
>         logger: logger.With("service", "mail"),
>     }
> }
> 
> func (s *mailService) SendUptimeReport(ctx context.Context, stats *UptimeStats, email string) error {
>     // Tạo HTML email template
>     htmlBody := s.generateUptimeReportHTML(stats)
>     
>     // Gửi email
>     return s.sendEmail(email, "Daily Uptime Report", htmlBody)
> }
> ```
> 
> ### 3. Cập nhật uptime.go:
> 
> ```go
> type UptimeService struct {
>     recordRepo  RecordRepository
>     mailService MailService
>     logger      logger.Logger
> }
> 
> func NewUptimeService(recordRepo RecordRepository, mailService MailService, logger logger.Logger) *UptimeService {
>     return &UptimeService{
>         recordRepo:  recordRepo,
>         mailService: mailService,
>         logger:      logger.With("service", "uptime"),
>     }
> }
> 
> func (s *UptimeService) GetUptimeStats(ctx context.Context, req *UptimeRequest) (*UptimeStats, error) {
>     return s.recordRepo.GetUptimeStats(ctx, req)
> }
> 
> func (s *UptimeService) SendUptimeReportByEmail(ctx context.Context, req *UptimeRequest) error {
>     // 1. Tính uptime stats
>     stats, err := s.GetUptimeStats(ctx, req)
>     if err != nil {
>         return fmt.Errorf("failed to get uptime stats: %w", err)
>     }
>     
>     // 2. Gửi email
>     email := req.Email
>     if email == "" {
>         // Lấy admin email từ config
>         email = config.Get().SMTP.AdminEmail
>     }
>     
>     return s.mailService.SendUptimeReport(ctx, stats, email)
> }
> ```
> 
> ### 4. Tạo `usecases/server/uptime_usecase.go`:
> 
> ```go
> package server
> 
> type UptimeUsecase struct {
>     uptimeService server.UptimeService
>     logger        logger.Logger
> }
> 
> func NewUptimeUsecase(uptimeService server.UptimeService, logger logger.Logger) *UptimeUsecase {
>     return &UptimeUsecase{
>         uptimeService: uptimeService,
>         logger:        logger.With("usecase", "uptime"),
>     }
> }
> 
> func (u *UptimeUsecase) SendUptimeReport(ctx context.Context, req server.UptimeRequest) error {
>     return u.uptimeService.SendUptimeReportByEmail(ctx, &req)
> }
> ```
> 
> ### 5. Thêm handler trong `api/handler/uptime.go`:
> 
> ```go
> package handler
> 
> type UptimeHandler struct {
>     usecase UptimeUsecase
>     logger  logger.Logger
> }
> 
> func (h *UptimeHandler) SendUptimeReport(c *gin.Context) {
>     var req server.UptimeRequest
>     if err := c.ShouldBindJSON(&req); err != nil {
>         c.JSON(400, gin.H{"error": "Invalid request"})
>         return
>     }
>     
>     if err := h.usecase.SendUptimeReport(c.Request.Context(), req); err != nil {
>         c.JSON(500, gin.H{"error": "Failed to send uptime report"})
>         return
>     }
>     
>     c.JSON(200, gin.H{"message": "Uptime report sent successfully"})
> }
> ```
> 
> ## Vị trí files cần tạo/cập nhật:
> 
> 1. **Cập nhật**: interface.go
> 2. **Cập nhật**: uptime.go
> 3. **Tạo mới**: mail.go
> 4. **Tạo mới**: `usecases/server/uptime_usecase.go`
> 5. **Tạo mới**: uptime.go
> 6. **Cập nhật**: route.go - thêm route mới
> 7. **Cập nhật**: main.go - wire dependencies
> 
> Cách tiếp cận này tuân thủ Clean Architecture, dễ test, và có thể mở rộng trong tương lai.
