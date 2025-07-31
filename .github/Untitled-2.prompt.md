
@workspace triển khai api report thông tin uptime server, output là email gửi đến admin

> Copilot: Dựa trên codebase hiện tại, tôi sẽ triển khai API report thông tin uptime server và gửi email đến admin. Đây là cách triển khai:
> 
> ## 1. Cập nhật interface UseCase
> 
> ````go
> // ...existing code...
> 
> type UseCase interface {
> 	CreateServer(ctx context.Context, req CreateServerRequest) (*entity.Server, error)
> 	ViewServer(ctx context.Context, req QueryServerRequest) (*QueryServerResponse, error)
> 	UpdateServer(ctx context.Context, req UpdateServerRequest) error
> 	DeleteServer(ctx context.Context, serverID string) error
> 
> 	ImportServersFromExcel(ctx context.Context, file multipart.File) (*ImportRespose, error)
> 	ExportServersToExcel(ctx context.Context, req QueryServerRequest) error
> 	
> 	// Thêm API report uptime
> 	GenerateUptimeReport(ctx context.Context, req UptimeReportRequest) error
> }
> 
> // Thêm các struct mới
> type UptimeReportRequest struct {
> 	StartDate time.Time `json:"start_date" validate:"required" form:"start_date"`
> 	EndDate   time.Time `json:"end_date" validate:"required" form:"end_date"`
> 	Email     string    `json:"email" validate:"omitempty,email" form:"email"` // Optional, sẽ dùng admin email nếu không có
> }
> 
> type UptimeReportResponse struct {
> 	Stats       *UptimeStats      `json:"stats"`
> 	ServerStats []*ServerUptime   `json:"server_stats"`
> 	GeneratedAt time.Time         `json:"generated_at"`
> }
> 
> type ServerUptime struct {
> 	ServerID         string  `json:"server_id"`
> 	ServerName       string  `json:"server_name"`
> 	CurrentStatus    string  `json:"current_status"`
> 	UptimePercentage float64 `json:"uptime_percentage"`
> 	TotalOnlineTime  float64 `json:"total_online_time_hours"`
> 	TotalDownTime    float64 `json:"total_downtime_hours"`
> 	LastChecked      time.Time `json:"last_checked"`
> }
> ````
> 
> ## 2. Tạo Email Service
> 
> ````go
> package email
> 
> import (
> 	"bytes"
> 	"fmt"
> 	"html/template"
> 	"time"
> 
> 	"github.com/lits-06/vcs-sms/config"
> 	"github.com/lits-06/vcs-sms/services/server"
> 	"gopkg.in/gomail.v2"
> )
> 
> type SMTPService struct {
> 	config *config.SMTPConfig
> }
> 
> func NewSMTPService(cfg *config.SMTPConfig) *SMTPService {
> 	return &SMTPService{
> 		config: cfg,
> 	}
> }
> 
> func (s *SMTPService) SendUptimeReport(report *server.UptimeReportResponse, email string) error {
> 	if email == "" {
> 		email = s.config.AdminEmail
> 	}
> 
> 	// Generate HTML content
> 	htmlContent, err := s.generateHTMLReport(report)
> 	if err != nil {
> 		return fmt.Errorf("failed to generate HTML report: %w", err)
> 	}
> 
> 	// Create email message
> 	m := gomail.NewMessage()
> 	m.SetHeader("From", s.config.From)
> 	m.SetHeader("To", email)
> 	m.SetHeader("Subject", fmt.Sprintf("Server Uptime Report - %s to %s", 
> 		report.Stats.StartDate.Format("2006-01-02"), 
> 		report.Stats.EndDate.Format("2006-01-02")))
> 	m.SetBody("text/html", htmlContent)
> 
> 	// Send email
> 	d := gomail.NewDialer(s.config.Host, s.config.Port, s.config.Username, s.config.Password)
> 	
> 	if err := d.DialAndSend(m); err != nil {
> 		return fmt.Errorf("failed to send email: %w", err)
> 	}
> 
> 	return nil
> }
> 
> func (s *SMTPService) generateHTMLReport(report *server.UptimeReportResponse) (string, error) {
> 	tmpl := `
> <!DOCTYPE html>
> <html>
> <head>
>     <meta charset="UTF-8">
>     <title>Server Uptime Report</title>
>     <style>
>         body { font-family: Arial, sans-serif; margin: 20px; }
>         .header { background-color: #f4f4f4; padding: 20px; border-radius: 5px; }
>         .stats { background-color: #e8f5e8; padding: 15px; margin: 20px 0; border-radius: 5px; }
>         .warning { background-color: #fff3cd; padding: 15px; margin: 20px 0; border-radius: 5px; }
>         .error { background-color: #f8d7da; padding: 15px; margin: 20px 0; border-radius: 5px; }
>         table { width: 100%; border-collapse: collapse; margin: 20px 0; }
>         th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
>         th { background-color: #f2f2f2; }
>         .online { color: #28a745; font-weight: bold; }
>         .offline { color: #dc3545; font-weight: bold; }
>         .good { background-color: #d4edda; }
>         .warning-row { background-color: #fff3cd; }
>         .bad { background-color: #f8d7da; }
>     </style>
> </head>
> <body>
>     <div class="header">
>         <h1>📊 Server Uptime Report</h1>
>         <p><strong>Period:</strong> {{.Stats.StartDate.Format "2006-01-02"}} to {{.Stats.EndDate.Format "2006-01-02"}}</p>
>         <p><strong>Generated At:</strong> {{.GeneratedAt.Format "2006-01-02 15:04:05 MST"}}</p>
>     </div>
> 
>     <div class="stats">
>         <h2>📈 Overall Statistics</h2>
>         <ul>
>             <li><strong>Total Servers:</strong> {{.Stats.TotalServers}}</li>
>             <li><strong>Online Servers:</strong> <span class="online">{{.Stats.OnlineServers}}</span></li>
>             <li><strong>Offline Servers:</strong> <span class="offline">{{.Stats.OfflineServers}}</span></li>
>             <li><strong>Average Uptime:</strong> {{printf "%.2f" .Stats.UptimePercentage}}%</li>
>         </ul>
>     </div>
> 
>     {{if lt .Stats.UptimePercentage 95.0}}
>     <div class="warning">
>         <h3>⚠️ Warning</h3>
>         <p>Overall uptime is below 95% ({{printf "%.2f" .Stats.UptimePercentage}}%). Please review server status.</p>
>     </div>
>     {{end}}
> 
>     <h2>🖥️ Individual Server Status</h2>
>     <table>
>         <thead>
>             <tr>
>                 <th>Server ID</th>
>                 <th>Server Name</th>
>                 <th>Current Status</th>
>                 <th>Uptime %</th>
>                 <th>Online Hours</th>
>                 <th>Downtime Hours</th>
>                 <th>Last Checked</th>
>             </tr>
>         </thead>
>         <tbody>
>             {{range .ServerStats}}
>             <tr class="{{if ge .UptimePercentage 99.0}}good{{else if ge .UptimePercentage 95.0}}warning-row{{else}}bad{{end}}">
>                 <td>{{.ServerID}}</td>
>                 <td>{{.ServerName}}</td>
>                 <td class="{{if eq .CurrentStatus "ON"}}online{{else}}offline{{end}}">
>                     {{if eq .CurrentStatus "ON"}}🟢 ONLINE{{else}}🔴 OFFLINE{{end}}
>                 </td>
>                 <td>{{printf "%.2f" .UptimePercentage}}%</td>
>                 <td>{{printf "%.1f" .TotalOnlineTime}}</td>
>                 <td>{{printf "%.1f" .TotalDownTime}}</td>
>                 <td>{{.LastChecked.Format "2006-01-02 15:04"}}</td>
>             </tr>
>             {{end}}
>         </tbody>
>     </table>
> 
>     <div class="footer" style="margin-top: 30px; padding: 20px; background-color: #f4f4f4; border-radius: 5px; font-size: 12px; color: #666;">
>         <p>This report was automatically generated by the Server Management System.</p>
>         <p>For technical support, please contact the system administrator.</p>
>     </div>
> </body>
> </html>
> `
> 
> 	t, err := template.New("uptime_report").Parse(tmpl)
> 	if err != nil {
> 		return "", err
> 	}
> 
> 	var buf bytes.Buffer
> 	if err := t.Execute(&buf, report); err != nil {
> 		return "", err
> 	}
> 
> 	return buf.String(), nil
> }
> ````
> 
> ## 3. Cập nhật Service Layer
> 
> ````go
> // ...existing code...
> 
> func (uc *ServerUsecase) GenerateUptimeReport(ctx context.Context, req UptimeReportRequest) error {
> 	// Validate date range
> 	if req.EndDate.Before(req.StartDate) {
> 		return fmt.Errorf("end date cannot be before start date")
> 	}
> 	
> 	if req.EndDate.Sub(req.StartDate) > 30*24*time.Hour {
> 		return fmt.Errorf("date range cannot exceed 30 days")
> 	}
> 
> 	// Get overall uptime stats from Elasticsearch
> 	uptimeReq := &UptimeRequest{
> 		StartDate: req.StartDate,
> 		EndDate:   req.EndDate,
> 		Email:     req.Email,
> 	}
> 
> 	stats, err := uc.recordRepo.GetUptimeStats(ctx, uptimeReq)
> 	if err != nil {
> 		return fmt.Errorf("failed to get uptime stats: %w", err)
> 	}
> 
> 	// Get detailed server stats
> 	serverStats, err := uc.getDetailedServerStats(ctx, req.StartDate, req.EndDate)
> 	if err != nil {
> 		return fmt.Errorf("failed to get detailed server stats: %w", err)
> 	}
> 
> 	// Prepare report
> 	report := &UptimeReportResponse{
> 		Stats:       stats,
> 		ServerStats: serverStats,
> 		GeneratedAt: time.Now(),
> 	}
> 
> 	// Send email
> 	emailService := uc.emailService // Inject this in constructor
> 	if err := emailService.SendUptimeReport(report, req.Email); err != nil {
> 		return fmt.Errorf("failed to send uptime report email: %w", err)
> 	}
> 
> 	return nil
> }
> 
> func (uc *ServerUsecase) getDetailedServerStats(ctx context.Context, startDate, endDate time.Time) ([]*ServerUptime, error) {
> 	// Get all servers
> 	allServers, _, err := uc.serverRepo.List(ctx, ServerFilter{}, ServerSort{}, ServerPagination{})
> 	if err != nil {
> 		return nil, fmt.Errorf("failed to get servers: %w", err)
> 	}
> 
> 	var serverStats []*ServerUptime
> 	totalHours := endDate.Sub(startDate).Hours()
> 
> 	for _, srv := range *allServers {
> 		// Get uptime data for this specific server from Elasticsearch
> 		serverUptime, err := uc.recordRepo.GetServerUptimeStats(ctx, srv.ID, startDate, endDate)
> 		if err != nil {
> 			// Log error but continue with other servers
> 			fmt.Printf("Failed to get uptime for server %s: %v\n", srv.ID, err)
> 			continue
> 		}
> 
> 		uptimePercentage := serverUptime.OnlineHours / totalHours * 100
> 		downTime := totalHours - serverUptime.OnlineHours
> 
> 		serverStats = append(serverStats, &ServerUptime{
> 			ServerID:         srv.ID,
> 			ServerName:       srv.Name,
> 			CurrentStatus:    string(srv.Status),
> 			UptimePercentage: uptimePercentage,
> 			TotalOnlineTime:  serverUptime.OnlineHours,
> 			TotalDownTime:    downTime,
> 			LastChecked:      srv.UpdatedAt,
> 		})
> 	}
> 
> 	return serverStats, nil
> }
> ````
> 
> ## 4. Cập nhật Elasticsearch Repository
> 
> ````go
> // ...existing code...
> 
> type ServerUptimeData struct {
> 	OnlineHours float64 `json:"online_hours"`
> }
> 
> func (r *elasticRepository) GetServerUptimeStats(ctx context.Context, serverID string, startDate, endDate time.Time) (*ServerUptimeData, error) {
> 	query := map[string]interface{}{
> 		"query": map[string]interface{}{
> 			"bool": map[string]interface{}{
> 				"must": []map[string]interface{}{
> 					{
> 						"term": map[string]interface{}{
> 							"server_id": serverID,
> 						},
> 					},
> 					{
> 						"range": map[string]interface{}{
> 							"timestamp": map[string]interface{}{
> 								"gte": startDate.Format(time.RFC3339),
> 								"lte": endDate.Format(time.RFC3339),
> 							},
> 						},
> 					},
> 				},
> 			},
> 		},
> 		"aggs": map[string]interface{}{
> 			"online_time": map[string]interface{}{
> 				"filter": map[string]interface{}{
> 					"term": map[string]interface{}{
> 						"status": "ON",
> 					},
> 				},
> 				"aggs": map[string]interface{}{
> 					"total_seconds": map[string]interface{}{
> 						"sum": map[string]interface{}{
> 							"field": "interval",
> 						},
> 					},
> 				},
> 			},
> 		},
> 		"size": 0,
> 	}
> 
> 	queryJSON, _ := json.Marshal(query)
> 	indices := r.getIndicesForTimeRange(startDate, endDate)
> 	indexPattern := strings.Join(indices, ",")
> 
> 	res, err := r.client.Search(
> 		r.client.Search.WithContext(ctx),
> 		r.client.Search.WithIndex(indexPattern),
> 		r.client.Search.WithBody(bytes.NewReader(queryJSON)),
> 	)
> 	if err != nil {
> 		return nil, fmt.Errorf("failed to search server uptime: %w", err)
> 	}
> 	defer res.Body.Close()
> 
> 	if res.IsError() {
> 		return nil, fmt.Errorf("Elasticsearch search error: %s", res.String())
> 	}
> 
> 	var result struct {
> 		Aggregations struct {
> 			OnlineTime struct {
> 				TotalSeconds struct {
> 					Value float64 `json:"value"`
> 				} `json:"total_seconds"`
> 			} `json:"online_time"`
> 		} `json:"aggregations"`
> 	}
> 
> 	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
> 		return nil, fmt.Errorf("failed to decode response: %w", err)
> 	}
> 
> 	onlineHours := result.Aggregations.OnlineTime.TotalSeconds.Value / 3600 // Convert seconds to hours
> 
> 	return &ServerUptimeData{
> 		OnlineHours: onlineHours,
> 	}, nil
> }
> ````
> 
> ## 5. Cập nhật Handler
> 
> ````go
> // ...existing code...
> 
> func (h *ServerHandler) GenerateUptimeReport(c *gin.Context) {
> 	var req server.UptimeReportRequest
> 	if err := c.ShouldBindQuery(&req); err != nil {
> 		h.logger.Error("Failed to bind query", "error", err)
> 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
> 		return
> 	}
> 
> 	// Validate date format if provided as strings
> 	if startDateStr := c.Query("start_date"); startDateStr != "" {
> 		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
> 			req.StartDate = startDate
> 		} else {
> 			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD"})
> 			return
> 		}
> 	}
> 
> 	if endDateStr := c.Query("end_date"); endDateStr != "" {
> 		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
> 			req.EndDate = endDate
> 		} else {
> 			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD"})
> 			return
> 		}
> 	}
> 
> 	// Default to last 7 days if dates not provided
> 	if req.StartDate.IsZero() {
> 		req.StartDate = time.Now().AddDate(0, 0, -7)
> 	}
> 	if req.EndDate.IsZero() {
> 		req.EndDate = time.Now()
> 	}
> 
> 	if err := h.service.GenerateUptimeReport(c.Request.Context(), req); err != nil {
> 		h.logger.Error("Failed to generate uptime report", "error", err)
> 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate uptime report"})
> 		return
> 	}
> 
> 	h.logger.Info("Uptime report generated successfully", 
> 		"start_date", req.StartDate.Format("2006-01-02"),
> 		"end_date", req.EndDate.Format("2006-01-02"),
> 		"email", req.Email)
> 	
> 	c.JSON(http.StatusOK, gin.H{
> 		"message": "Uptime report has been generated and sent via email",
> 		"start_date": req.StartDate.Format("2006-01-02"),
> 		"end_date": req.EndDate.Format("2006-01-02"),
> 	})
> }
> ````
> 
> ## 6. Cập nhật Router
> 
> ````go
> // ...existing code...
> 
> func (r *Route) SetupRoutes() *gin.Engine {
> 	router := gin.Default()
> 
> 	// Health check endpoint
> 	router.GET("/health", func(c *gin.Context) {
> 		c.JSON(200, gin.H{
> 			"status":  "ok",
> 			"message": "Server Management System is running",
> 		})
> 	})
> 
> 	// API v1 group
> 	v1 := router.Group("/api/v1")
> 	{
> 		// Server management routes
> 		servers := v1.Group("/servers")
> 		{
> 			// CRUD operations
> 			servers.POST("", r.serverHandler.CreateServer)
> 			servers.GET("", r.serverHandler.ViewServer)
> 			servers.PUT("/:id", r.serverHandler.UpdateServer)
> 			servers.DELETE("/:id", r.serverHandler.DeleteServer)
> 
> 			// Import/Export operations
> 			servers.POST("/import", r.serverHandler.ImportServersFromExcel)
> 			servers.GET("/export", r.serverHandler.ExportServersToExcel)
> 			
> 			// Report operations
> 			servers.POST("/report/uptime", r.serverHandler.GenerateUptimeReport)
> 		}
> 	}
> 
> 	return router
> }
> ````
> 
> ## 7. Cập nhật cấu hình SMTP trong config
> 
> ````yaml
> # ...existing code...
> 
> smtp:
>   host: smtp.gmail.com
>   port: 587
>   username: your-email@gmail.com
>   password: your-app-password
>   from: noreply@yourcompany.com
>   admin_email: admin@yourcompany.com
> ````
> 
> ## Cách sử dụng API:
> 
> ```bash
> # Generate uptime report for last 7 days (default)
> POST /api/v1/servers/report/uptime
> 
> # Generate uptime report for specific date range
> POST /api/v1/servers/report/uptime?start_date=2024-01-01&end_date=2024-01-07
> 
> # Send report to specific email
> POST /api/v1/servers/report/uptime?start_date=2024-01-01&end_date=2024-01-07&email=manager@company.com
> ```
> 
> ## Ưu điểm của cách triển khai này:
> 
> 1. **Tự động hóa hoàn toàn**: API tự động tạo report và gửi email
> 2. **Giao diện đẹp**: Email HTML với styling chuyên nghiệp và màu sắc phân loại
> 3. **Thông tin chi tiết**: Bao gồm cả thống kê tổng quan và chi tiết từng server
> 4. **Linh hoạt**: Có thể chọn khoảng thời gian và email nhận
> 5. **Cảnh báo thông minh**: Tự động highlight các server có uptime thấp
> 6. **Bảo mật**: Validate input và giới hạn khoảng thời gian query
> 7. **Performance**: Sử dụng Elasticsearch để tính toán uptime nhanh chóng
> 
> Triển khai này giúp admin có thể dễ dàng theo dõi tình trạng server và nhận report định kỳ qua email một cách chuyên nghiệp.
