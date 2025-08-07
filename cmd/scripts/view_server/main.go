package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/lits-06/vcs-sms/entity"
	"github.com/lits-06/vcs-sms/services/server"
)

var (
	serverTypes  = []string{"Web", "Database", "Cache", "Load Balancer", "API", "File", "Mail", "DNS", "Backup", "Monitor"}
	environments = []string{"Production", "Staging", "Development", "Testing", "Demo"}
	locations    = []string{"US-East", "US-West", "EU-Central", "Asia-Pacific", "Canada", "Australia"}
)

// Command line flags with detailed descriptions
var (
	filterMode = flag.String("filter", "blank",
		"Filter mode strategy:\n"+
			"  • full      - All filters (types, environments, locations, partials)\n"+
			"  • blank     - No name filters (only status/pagination)\n"+
			"  • partial   - Partial string matches only\n"+
			"  • env-only  - Environment filters only\n"+
			"  • type-only - Server type filters only\n"+
			"  • location-only - Location filters only")

	// runs = flag.Int("runs", 1,
	// 	"Number of test runs to execute (1-100)")

	// baseURL = flag.String("url", "http://localhost:8080",
	// 	"Base API URL (without /api/v1/servers path)")

	status = flag.String("status", "",
		"Force specific status filter (ON, OFF, or empty for random)")

	verbose = flag.Bool("v", false,
		"Verbose output (show detailed request info)")

	help = flag.Bool("help", false,
		"Show detailed help and usage examples")

	// seed = flag.Int64("seed", 0,
	// 	"Random seed for reproducible results (0 = use current time)")

	// showStats = flag.Bool("stats", true,
	// 	"Show summary statistics after results")
)

var (
	filterNames    = []string{}
	filterStatuses = []string{"", "ON", "OFF"}
	// filterIPs      = []string{"", "192.168", "10.", "172.16"}
	sortFields = []string{"", "name", "status", "created_at", "updated_at"}
	sortOrders = []string{"", "asc", "desc"}
)

func buildFullFilterNames() []string {
	var filters []string

	// Add empty filter for "no filter" case
	filters = append(filters, "")

	// Add server types (exact matches)
	filters = append(filters, serverTypes...)

	// Add environments for filtering servers by environment
	filters = append(filters, environments...)

	// Add locations for filtering servers by location
	filters = append(filters, locations...)

	return filters
}

func buildBlankFilterNames() []string {
	return []string{""}
}

func main() {
	flag.Usage = showUsage
	flag.Parse()

	if *help {
		showDetailedHelp()
		return
	}

	// Validate inputs
	// if err := validateFlags(); err != nil {
	// 	fmt.Fprintf(os.Stderr, "❌ Error: %v\n\n", err)
	// 	flag.Usage()
	// 	os.Exit(1)
	// }

	initializeFilters()

	fmt.Println("=== Random Server View Tool ===")
	fmt.Println()

	// Generate random parameters
	params := generateRandomParams()

	// Print parameters being used
	printViewParams(params)

	// Call view API and get results
	response := viewServers(params)

	// Print results
	printViewResults(response)
}

func showUsage() {
	fmt.Fprintf(os.Stderr, `
🔍 Server View Testing Tool

USAGE:
  %s [FLAGS]

FLAGS:
`, os.Args[0])
	flag.PrintDefaults()

	fmt.Fprintf(os.Stderr, `
EXAMPLES:
  %s                           # Default: full filtering, single run
  %s -filter=blank -runs=5     # Test without name filters, 5 runs
  %s -filter=env-only -v       # Environment filtering with verbose output
  %s -url=http://prod:8080     # Test against production server
  %s -help                     # Show detailed help

ENVIRONMENT VARIABLES:
  FILTER_MODE    Override filter mode (same values as -filter flag)

`, os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}

func showDetailedHelp() {
	fmt.Print(`
🔍 Server View Testing Tool - Detailed Help
═══════════════════════════════════════════

OVERVIEW:
  This tool tests the server view API by generating random filter, sort,
  and pagination parameters. It helps verify API functionality across
  different scenarios.

FILTER MODES:
  ┌──────────────┬─────────────────────────────────────────────────────┐
  │ Mode         │ Description                                         │
  ├──────────────┼─────────────────────────────────────────────────────┤
  │ full         │ All available filters: server types, environments,  │
  │              │ locations, and partial string matches               │
  │              │ Examples: "Web", "Production", "US-East", "Server"  │
  ├──────────────┼─────────────────────────────────────────────────────┤
  │ blank        │ No name filtering, only status and pagination       │
  │              │ Useful for testing base functionality               │
  ├──────────────┼─────────────────────────────────────────────────────┤
  │ partial      │ Partial string matches only                         │
  │              │ Examples: "Server", "01", "Load", "Data"            │
  ├──────────────┼─────────────────────────────────────────────────────┤
  │ env-only     │ Environment filters only                            │
  │              │ Examples: "Production", "Staging", "Development"    │
  ├──────────────┼─────────────────────────────────────────────────────┤
  │ type-only    │ Server type filters only                            │
  │              │ Examples: "Web", "Database", "API", "Cache"         │
  ├──────────────┼─────────────────────────────────────────────────────┤
  │ location-only│ Location filters only                               │
  │              │ Examples: "US-East", "EU-Central", "Asia-Pacific"   │
  └──────────────┴─────────────────────────────────────────────────────┘

USAGE EXAMPLES:

  🚀 Quick Testing:
	go run main.go                    # Single test with full filtering
	go run main.go -filter=blank      # Test without name filters

  📊 Multiple Runs:
	go run main.go -runs=10           # Run 10 tests for consistency
	go run main.go -runs=5 -v         # 5 runs with verbose output

  🎯 Specific Filter Testing:
	go run main.go -filter=env-only -runs=3    # Test environment filtering
	go run main.go -filter=type-only -v        # Test server type filtering

  🌐 Different Environments:
	go run main.go -url=http://staging:8080    # Test staging environment
	go run main.go -url=https://api.prod.com   # Test production API

  🔍 Debugging:
	go run main.go -v -seed=12345              # Verbose with fixed seed
	go run main.go -filter=partial -runs=1 -v # Debug partial filtering

  📈 Performance Testing:
	go run main.go -runs=50 -stats=false      # Many runs, no detailed stats

ENVIRONMENT VARIABLES:
  FILTER_MODE=env-only go run main.go         # Override filter mode
  
RANDOM PARAMETERS GENERATED:
  • Name filters (based on selected mode)
  • Status filters (ON, OFF, or empty)
  • Sort fields (name, status, created_at, updated_at)
  • Sort orders (asc, desc)
  • Pagination (random from/to ranges)

OUTPUT EXPLANATION:
  🎲 Parameters section shows what filters were randomly selected
  📋 API call section shows the actual URL being called
  📊 Results section shows servers found and detailed statistics
  📈 Summary section shows distribution analysis

TIPS:
  • Use -v flag to see actual API URLs for debugging
  • Use -seed flag to reproduce specific test scenarios
  • Use different -runs values to test consistency
  • Combine with shell scripts for automated testing

RETURN CODES:
  0  Success
  1  Invalid arguments or API errors
`)
}

// func validateFlags() error {
// 	// Validate runs
// 	if *runs < 1 || *runs > 100 {
// 		return fmt.Errorf("runs must be between 1 and 100, got %d", *runs)
// 	}

// 	// Validate filter mode
// 	validModes := []string{"full", "blank", "partial", "env-only", "type-only", "location-only", "complete", "empty", "none", "environment", "types", "locations", "part"}
// 	valid := false
// 	for _, mode := range validModes {
// 		if strings.ToLower(*filterMode) == mode {
// 			valid = true
// 			break
// 		}
// 	}
// 	if !valid {
// 		return fmt.Errorf("invalid filter mode '%s'. Valid options: full, blank, partial, env-only, type-only, location-only", *filterMode)
// 	}

// 	// Validate URL format
// 	if !strings.HasPrefix(*baseURL, "http://") && !strings.HasPrefix(*baseURL, "https://") {
// 		return fmt.Errorf("base URL must start with http:// or https://, got '%s'", *baseURL)
// 	}

// 	return nil
// }

func initializeFilters() {
	// Check environment variable first
	if envMode := os.Getenv("FILTER_MODE"); envMode != "" {
		*filterMode = envMode
	}

	if *verbose {
		fmt.Printf("🔧 Initializing filters in '%s' mode\n", *filterMode)
	}

	switch strings.ToLower(*filterMode) {
	case "full", "complete":
		filterNames = buildFullFilterNames()
		if *verbose {
			fmt.Printf("   📝 Full filters: %d options available\n", len(filterNames))
		}
	case "blank", "empty", "none":
		filterNames = buildBlankFilterNames()
		if *verbose {
			fmt.Printf("   📝 Blank filters: only empty filter available\n")
		}
		// case "partial", "part":
		// 	filterNames = buildPartialFilterNames()
		// 	if *verbose {
		// 		fmt.Printf("   📝 Partial filters: %d partial match options\n", len(filterNames))
		// 	}
		// case "env-only", "environment":
		// 	filterNames = buildEnvironmentOnlyFilters()
		// 	if *verbose {
		// 		fmt.Printf("   📝 Environment filters: %d environments\n", len(filterNames))
		// 	}
		// case "type-only", "types":
		// 	filterNames = buildTypeOnlyFilters()
		// 	if *verbose {
		// 		fmt.Printf("   📝 Type filters: %d server types\n", len(filterNames))
		// 	}
		// case "location-only", "locations":
		// 	filterNames = buildLocationOnlyFilters()
		// 	if *verbose {
		// 		fmt.Printf("   📝 Location filters: %d locations\n", len(filterNames))
		// 	}
	}

	if *verbose {
		fmt.Println()
	}
}

func generateRandomParams() url.Values {
	params := url.Values{}

	// Random filter by name
	filterName := filterNames[rand.Intn(len(filterNames))]
	if filterName != "" {
		params.Add("name", filterName)
	}

	// Random filter by status
	var filterStatus string
	if *status != "" {
		filterStatus = *status
		if *verbose {
			fmt.Printf("🎯 Using specified status: %s\n", filterStatus)
		}
	} // else {
	// 	filterStatus = filterStatuses[rand.Intn(len(filterStatuses))]
	// }

	if filterStatus != "" {
		params.Add("status", filterStatus)
	}

	// Random filter by IPv4
	// filterIP := filterIPs[rand.Intn(len(filterIPs))]
	// if filterIP != "" {
	// 	params.Add("ipv4", filterIP)
	// }

	// Random sort
	sortField := sortFields[rand.Intn(len(sortFields))]
	sortOrder := sortOrders[rand.Intn(len(sortOrders))]
	if sortField != "" {
		params.Add("sort", sortField)
	}
	if sortOrder != "" {
		params.Add("order", sortOrder)
	}

	// Random pagination
	// from := rand.Intn(5)           // Start from 0-4
	// to := from + rand.Intn(20) + 5 // Get 5-25 records
	// params.Add("from", fmt.Sprintf("%d", from))
	// params.Add("to", fmt.Sprintf("%d", to))

	return params
}

func printViewParams(params url.Values) {
	fmt.Println("🎲 Random View Parameters:")
	fmt.Println("─────────────────────────────")

	if name := params.Get("name"); name != "" {
		fmt.Printf("  📛 Filter by name: %s\n", name)
	}

	if status := params.Get("status"); status != "" {
		statusIcon := "🔴"
		if status == "ON" {
			statusIcon = "🟢"
		}
		fmt.Printf("  %s Filter by status: %s\n", statusIcon, status)
	}

	if ipv4 := params.Get("ipv4"); ipv4 != "" {
		fmt.Printf("  🌐 Filter by IPv4: %s*\n", ipv4)
	}

	if sort := params.Get("sort"); sort != "" {
		order := params.Get("order")
		if order == "" {
			order = "asc"
		}
		arrow := "⬆️"
		if order == "desc" {
			arrow = "⬇️"
		}
		fmt.Printf("  %s Sort by: %s (%s)\n", arrow, sort, order)
	}

	from := params.Get("from")
	to := params.Get("to")
	if from != "" && to != "" {
		fmt.Printf("  📄 Pagination: from %s to %s\n", from, to)
	}

	if len(params) == 0 {
		fmt.Println("  No filters applied (view all)")
	}

	fmt.Println()
}

func viewServers(params url.Values) *server.QueryServerResponse {
	baseURL := "http://localhost:8080/api/v1/servers"

	queryURL := baseURL
	if len(params) > 0 {
		queryURL += "?" + params.Encode()
	}

	fmt.Printf("📋 Calling view API: %s\n", queryURL)
	fmt.Println()

	resp, err := http.Get(queryURL)
	if err != nil {
		log.Fatalf("Error calling view API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("View API failed (Status: %d): %s", resp.StatusCode, string(body))
	}

	var response server.QueryServerResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Fatalf("Error decoding response: %v", err)
	}

	return &response
}

func printViewResults(response *server.QueryServerResponse) {
	fmt.Println("📊 View Results:")
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════════")

	if response.Servers == nil || len(*response.Servers) == 0 {
		fmt.Println("❌ No servers found matching the criteria")
		return
	}

	fmt.Printf("📈 Total servers found: %d\n", response.Total)
	fmt.Printf("📋 Servers returned: %d\n\n", len(*response.Servers))

	// Print header
	fmt.Printf("%-12s %-35s %-15s %-8s %-20s %-20s\n",
		"ID", "NAME", "IPv4", "STATUS", "CREATED AT", "UPDATED AT")
	fmt.Println("─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────")

	// Print servers
	for i, server := range *response.Servers {
		// Truncate long names
		name := server.Name
		if len(name) > 33 {
			name = name[:30] + "..."
		}

		// Format timestamps
		createdAt := server.CreatedAt.Format("2006-01-02 15:04")
		updatedAt := server.UpdatedAt.Format("2006-01-02 15:04")

		// Status with color indicators
		statusIcon := "🔴"
		if server.Status == "ON" {
			statusIcon = "🟢"
		}

		fmt.Printf("%-12s %-35s %-15s %s %-6s %-20s %-20s\n",
			server.ID,
			name,
			server.IPv4,
			statusIcon,
			server.Status,
			createdAt,
			updatedAt)

		// Add separator every 10 rows for better readability
		if (i+1)%10 == 0 && i+1 < len(*response.Servers) {
			fmt.Println("┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈")
		}
	}

	// Summary statistics
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════")
	printSummaryStats(response.Servers, response.Total)
}

func printSummaryStats(servers *[]entity.Server, total int) {
	onCount := 0
	offCount := 0
	ipPrefixes := make(map[string]int)
	nameTypes := make(map[string]int)

	for _, server := range *servers {
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

		// Count server types from names
		if len(server.Name) > 0 {
			// Extract first word as server type
			parts := []string{}
			currentWord := ""
			for _, char := range server.Name {
				if char == ' ' {
					if currentWord != "" {
						parts = append(parts, currentWord)
						break
					}
				} else {
					currentWord += string(char)
				}
			}
			if currentWord != "" && len(parts) == 0 {
				parts = append(parts, currentWord)
			}

			if len(parts) > 0 {
				serverType := parts[0]
				nameTypes[serverType]++
			}
		}
	}

	fmt.Println("📊 Summary Statistics:")

	if len(*servers) > 0 {
		fmt.Printf("  🟢 Online servers:  %d (%.1f%%)\n", onCount, float64(onCount)/float64(len(*servers))*100)
		fmt.Printf("  🔴 Offline servers: %d (%.1f%%)\n", offCount, float64(offCount)/float64(len(*servers))*100)
	}

	if total > len(*servers) {
		fmt.Printf("  📄 Showing %d of %d total servers\n", len(*servers), total)
	}

	if len(ipPrefixes) > 0 {
		fmt.Println("  🌐 Network distribution:")
		for prefix, count := range ipPrefixes {
			fmt.Printf("     %s*: %d servers\n", prefix, count)
		}
	}

	if len(nameTypes) > 0 {
		fmt.Println("  🏷️  Server types:")
		for serverType, count := range nameTypes {
			fmt.Printf("     %s: %d servers\n", serverType, count)
		}
	}

	fmt.Printf("\n✅ View completed at: %s\n", time.Now().Format("2006-01-02 15:04:05"))
}
