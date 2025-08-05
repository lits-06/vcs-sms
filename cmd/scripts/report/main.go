package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/lits-06/vcs-sms/services/server"
)

func main() {
	baseURL := "http://localhost:8080/api/v1/reports/uptime"

	fmt.Println("🚀 Starting Uptime Report API Testing")
	fmt.Println("═══════════════════════════════════════")

	req := generateRandomUptimeRequest()

	// Hiển thị thông tin request
	printRequestInfo(req)

	// Gọi API

	callUptimeReportAPI(baseURL, req)
}

func generateRandomUptimeRequest() server.UptimeRequest {
	// Random date range (trong 30 ngày gần đây)
	now := time.Now()

	// // Start date: từ 30 ngày trước đến 7 ngày trước
	// daysBack := rand.Intn(23) + 7 // 7-30 ngày
	// startDate := now.AddDate(0, 0, -daysBack)

	// // End date: từ start date đến hôm nay
	// daysDuration := rand.Intn(daysBack) + 1 // ít nhất 1 ngày
	// endDate := startDate.AddDate(0, 0, daysDuration)

	startDate := now
	endDate := now

	// Đảm bảo end date không vượt quá hôm nay
	if endDate.After(now) {
		endDate = now
	}

	return server.UptimeRequest{
		StartDate: startDate,
		EndDate:   endDate,
	}
}

func printRequestInfo(req server.UptimeRequest) {
	fmt.Println("📊 Random Uptime Report Request:")
	fmt.Println("─────────────────────────────────────")

	fmt.Printf("  📅 Date Range:\n")
	fmt.Printf("     Start: %s\n", req.StartDate.Format("2006-01-02"))
	fmt.Printf("     End:   %s\n", req.EndDate.Format("2006-01-02"))

	fmt.Println()
}

func callUptimeReportAPI(baseURL string, req server.UptimeRequest) {
	// Convert request to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		log.Fatalf("❌ Error marshaling request: %v", err)
	}

	// Create HTTP request
	resp, err := http.Post(baseURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("❌ Error creating request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("❌ API call failed (Status: %d)", resp.StatusCode)
	}

	fmt.Printf("✅ Report completed at: %s\n", time.Now().Format("2006-01-02 15:04:05"))
}
