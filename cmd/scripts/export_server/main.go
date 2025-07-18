package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

type Server struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	IPv4        string    `json:"ipv4"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	LastChecked time.Time `json:"last_checked"`
}

type ExportResponse struct {
	Servers []Server `json:"servers"`
	Total   int      `json:"total"`
}

var (
	filterNames    = []string{"", "Web", "Database", "API", "Cache", "Load Balancer", "File", "Mail"}
	filterStatuses = []string{"", "ON", "OFF"}
	filterIPs      = []string{"", "192.168", "10.", "172.16"}
	sortFields     = []string{"", "name", "status", "created_at", "updated_at"}
	sortOrders     = []string{"", "asc", "desc"}
)

func main() {
	// Seed random
	rand.Seed(time.Now().UnixNano())

	fmt.Println("=== Random Server Export Tool ===")
	fmt.Println()

	// Generate random parameters
	params := generateRandomParams()

	// Print parameters being used
	printExportParams(params)

	// Call export API and get results
	servers := exportAndGetServers(params)

	// Print results
	printExportResults(servers)
}

func generateRandomParams() url.Values {
	params := url.Values{}

	// Random filter by name
	filterName := filterNames[rand.Intn(len(filterNames))]
	if filterName != "" {
		params.Add("name", filterName)
	}

	// Random filter by status
	filterStatus := filterStatuses[rand.Intn(len(filterStatuses))]
	if filterStatus != "" {
		params.Add("status", filterStatus)
	}

	// Random filter by IPv4
	filterIP := filterIPs[rand.Intn(len(filterIPs))]
	if filterIP != "" {
		params.Add("ipv4", filterIP)
	}

	// Random sort
	sortField := sortFields[rand.Intn(len(sortFields))]
	sortOrder := sortOrders[rand.Intn(len(sortOrders))]
	params.Add("sort", sortField)
	params.Add("order", sortOrder)

	// Random pagination
	from := rand.Intn(10)           // Start from 0-9
	to := from + rand.Intn(50) + 10 // Get 10-60 records
	params.Add("from", fmt.Sprintf("%d", from))
	params.Add("to", fmt.Sprintf("%d", to))

	return params
}

func printExportParams(params url.Values) {
	fmt.Println("🎲 Random Export Parameters:")
	fmt.Println("─────────────────────────────")

	for key, values := range params {
		if len(values) > 0 {
			fmt.Printf("  %s: %s\n", key, values[0])
		}
	}

	if len(params) == 0 {
		fmt.Println("  No filters applied (export all)")
	}

	fmt.Println()
}

func exportAndGetServers(params url.Values) *[]Server {
	// First call the actual export API
	exportURL := "http://localhost:8080/api/v1/servers/export"
	if len(params) > 0 {
		exportURL += "?" + params.Encode()
	}

	fmt.Printf("📤 Calling export API: %s\n", exportURL)

	resp, err := http.Get(exportURL)
	if err != nil {
		log.Fatalf("Error calling export API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 204 {
		fmt.Println("✅ Excel file exported successfully!")
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ Export failed (Status: %d): %s\n", resp.StatusCode, string(body))
	}

	fmt.Println()

	// Now get the data for terminal display
	queryURL := "http://localhost:8080/api/v1/servers"
	if len(params) > 0 {
		queryURL += "?" + params.Encode()
	}

	fmt.Printf("📋 Fetching data for terminal display...\n")

	queryResp, err := http.Get(queryURL)
	if err != nil {
		log.Printf("Error fetching servers for display: %v", err)
		return &[]Server{}
	}
	defer queryResp.Body.Close()

	if queryResp.StatusCode != 200 {
		body, _ := io.ReadAll(queryResp.Body)
		log.Printf("Failed to fetch servers (Status: %d): %s", queryResp.StatusCode, string(body))
		return &[]Server{}
	}

	var response struct {
		Servers *[]Server `json:"servers"`
		Total   int       `json:"total"`
	}

	body, _ := io.ReadAll(queryResp.Body)
	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("Error parsing response: %v", err)
		return &[]Server{}
	}

	return response.Servers
}

func printExportResults(servers *[]Server) {
	fmt.Println("📊 Export Results:")
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════════")

	if len(*servers) == 0 {
		fmt.Println("❌ No servers found matching the criteria")
		return
	}

	fmt.Printf("📈 Total servers exported: %d\n\n", len(*servers))

	// Print header
	fmt.Printf("%-12s %-35s %-15s %-8s %-20s\n",
		"ID", "NAME", "IPv4", "STATUS", "CREATED AT")
	fmt.Println("─────────────────────────────────────────────────────────────────────────────────────────")

	// Print servers
	for i, server := range *servers {
		// Truncate long names
		name := server.Name
		if len(name) > 33 {
			name = name[:30] + "..."
		}

		// Format created time
		createdAt := server.CreatedAt.Format("2006-01-02 15:04")

		// Status with color indicators
		statusIcon := "🔴"
		if server.Status == "ON" {
			statusIcon = "🟢"
		}

		fmt.Printf("%-12s %-35s %-15s %s %-6s %-20s\n",
			server.ID,
			name,
			server.IPv4,
			statusIcon,
			server.Status,
			createdAt)

		// Add separator every 10 rows for better readability
		if (i+1)%10 == 0 && i+1 < len(*servers) {
			fmt.Println("┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈")
		}
	}

	// Summary statistics
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════════")
	printSummaryStats(*servers)
}

func printSummaryStats(servers []Server) {
	onCount := 0
	offCount := 0
	ipPrefixes := make(map[string]int)

	for _, server := range servers {
		if server.Status == "ON" {
			onCount++
		} else {
			offCount++
		}

		// Count IP prefixes
		if len(server.IPv4) >= 7 {
			prefix := server.IPv4[:7] // e.g., "192.168", "10.0.0.", etc.
			ipPrefixes[prefix]++
		}
	}

	fmt.Println("📊 Summary Statistics:")
	fmt.Printf("  🟢 Online servers:  %d (%.1f%%)\n", onCount, float64(onCount)/float64(len(servers))*100)
	fmt.Printf("  🔴 Offline servers: %d (%.1f%%)\n", offCount, float64(offCount)/float64(len(servers))*100)

	if len(ipPrefixes) > 0 {
		fmt.Println("  🌐 Network distribution:")
		for prefix, count := range ipPrefixes {
			fmt.Printf("     %s*: %d servers\n", prefix, count)
		}
	}

	fmt.Printf("\n✅ Export completed at: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("📁 Excel file saved in exports/ directory")
}
