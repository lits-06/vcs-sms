package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/lits-06/vcs-sms/pkg/utils"
	"github.com/lits-06/vcs-sms/services/server"
)

func main() {
	baseURL := "http://localhost:8080/api/v1/servers/import"

	excelFile := findExcelFile()
	if excelFile == "" {
		fmt.Println("❌ No Excel file found.")
		return
	}

	fmt.Println("🚀 Starting Server Import API Testing")
	fmt.Println("═══════════════════════════════════════")
	fmt.Printf("📁 Excel file: %s\n", excelFile)
	fmt.Printf("🌐 API URL: %s\n", baseURL)
	fmt.Println()

	response, err := callImportAPI(baseURL, excelFile)
	if err != nil {
		fmt.Printf("❌ Import failed: %v\n", err)
		return
	}

	// Display results
	displayImportResults(response)
}

func findExcelFile() string {
	projectRoot, err := utils.FindProjectRoot()
	if err != nil {
		fmt.Printf("❌ Failed to find project root: %v\n", err)
		return ""
	}

	// Tạo đường dẫn exports từ project root
	pattern := filepath.Join(projectRoot, "exports", "*.xlsx")

	// Tìm file Excel trong thư mục hiện tại
	files, err := filepath.Glob(pattern)
	if err != nil || len(files) == 0 {
		return ""
	}
	return files[0]
}

func callImportAPI(baseURL, filename string) (*server.ImportResponse, error) {
	// Open file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add file field
	fileWriter, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	writer.Close()

	// Create request
	req, err := http.NewRequest("POST", baseURL, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	fmt.Printf("📤 Uploading file: %s\n", filename)
	fmt.Printf("⏳ Calling import API...\n")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("import API failed (Status: %d): %s", resp.StatusCode, string(body))
	}

	var response server.ImportResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

func displayImportResults(response *server.ImportResponse) {
	fmt.Println()
	fmt.Println("📊 Import Results:")
	fmt.Println("═══════════════════════════════════════════════════════════════")

	// Summary
	fmt.Printf("✅ Success: %d servers\n", response.SuccessCount)
	fmt.Printf("❌ Failed:  %d servers\n", response.FailureCount)
	fmt.Printf("📈 Total:   %d servers\n", response.SuccessCount+response.FailureCount)

	if response.SuccessCount > 0 {
		fmt.Println()
		fmt.Println("✅ Successfully Imported Servers:")
		fmt.Println("─────────────────────────────────────────")
		for i, server := range response.SuccessServers {
			fmt.Printf("  %d. %s\n", i+1, server)
		}
	}

	if response.FailureCount > 0 {
		fmt.Println()
		fmt.Println("❌ Failed to Import Servers:")
		fmt.Println("─────────────────────────────────────────")
		for i, server := range response.FailureServers {
			fmt.Printf("  %d. %s\n", i+1, server)
		}
	}

	// Success rate
	total := response.SuccessCount + response.FailureCount
	if total > 0 {
		successRate := float64(response.SuccessCount) / float64(total) * 100
		fmt.Println()
		fmt.Printf("📊 Success Rate: %.1f%% (%d/%d)\n", successRate, response.SuccessCount, total)
	}

	fmt.Printf("\n✅ Import completed at: %s\n", time.Now().Format("2006-01-02 15:04:05"))
}
