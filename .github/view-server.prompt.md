
@workspace tạo script gọi api view server

> Copilot: Tôi sẽ tạo script để gọi API view server với các tính năng filter, pagination và sort ngẫu nhiên.
> 
> ````go
> package main
> 
> import (
> 	"encoding/json"
> 	"fmt"
> 	"io"
> 	"log"
> 	"math/rand"
> 	"net/http"
> 	"net/url"
> 	"time"
> )
> 
> type Server struct {
> 	ID        string    `json:"server_id"`
> 	Name      string    `json:"name"`
> 	IPv4      string    `json:"ipv4"`
> 	Status    string    `json:"status"`
> 	CreatedAt time.Time `json:"created_at"`
> 	UpdatedAt time.Time `json:"updated_at"`
> }
> 
> type ViewServerResponse struct {
> 	Servers *[]Server `json:"servers"`
> 	Total   int       `json:"total"`
> }
> 
> var (
> 	filterNames    = []string{"", "Web", "Database", "API", "Cache", "Load Balancer", "File", "Mail"}
> 	filterStatuses = []string{"", "ON", "OFF"}
> 	filterIPs      = []string{"", "192.168", "10.", "172.16"}
> 	sortFields     = []string{"", "name", "status", "created_at", "updated_at"}
> 	sortOrders     = []string{"", "asc", "desc"}
> )
> 
> func main() {
> 	// Seed random
> 	rand.Seed(time.Now().UnixNano())
> 
> 	fmt.Println("=== Random Server View Tool ===")
> 	fmt.Println()
> 
> 	// Generate random parameters
> 	params := generateRandomParams()
> 
> 	// Print parameters being used
> 	printViewParams(params)
> 
> 	// Call view API and get results
> 	response := viewServers(params)
> 
> 	// Print results
> 	printViewResults(response)
> }
> 
> func generateRandomParams() url.Values {
> 	params := url.Values{}
> 
> 	// Random filter by name
> 	filterName := filterNames[rand.Intn(len(filterNames))]
> 	if filterName != "" {
> 		params.Add("name", filterName)
> 	}
> 
> 	// Random filter by status
> 	filterStatus := filterStatuses[rand.Intn(len(filterStatuses))]
> 	if filterStatus != "" {
> 		params.Add("status", filterStatus)
> 	}
> 
> 	// Random filter by IPv4
> 	filterIP := filterIPs[rand.Intn(len(filterIPs))]
> 	if filterIP != "" {
> 		params.Add("ipv4", filterIP)
> 	}
> 
> 	// Random sort
> 	sortField := sortFields[rand.Intn(len(sortFields))]
> 	sortOrder := sortOrders[rand.Intn(len(sortOrders))]
> 	if sortField != "" {
> 		params.Add("sort", sortField)
> 	}
> 	if sortOrder != "" {
> 		params.Add("order", sortOrder)
> 	}
> 
> 	// Random pagination
> 	from := rand.Intn(10)           // Start from 0-9
> 	to := from + rand.Intn(20) + 5  // Get 5-25 records
> 	params.Add("from", fmt.Sprintf("%d", from))
> 	params.Add("to", fmt.Sprintf("%d", to))
> 
> 	return params
> }
> 
> func printViewParams(params url.Values) {
> 	fmt.Println("🎲 Random View Parameters:")
> 	fmt.Println("─────────────────────────────")
> 
> 	if name := params.Get("name"); name != "" {
> 		fmt.Printf("  📛 Filter by name: %s\n", name)
> 	}
> 	
> 	if status := params.Get("status"); status != "" {
> 		statusIcon := "🔴"
> 		if status == "ON" {
> 			statusIcon = "🟢"
> 		}
> 		fmt.Printf("  %s Filter by status: %s\n", statusIcon, status)
> 	}
> 	
> 	if ipv4 := params.Get("ipv4"); ipv4 != "" {
> 		fmt.Printf("  🌐 Filter by IPv4: %s*\n", ipv4)
> 	}
> 	
> 	if sort := params.Get("sort"); sort != "" {
> 		order := params.Get("order")
> 		if order == "" {
> 			order = "asc"
> 		}
> 		arrow := "⬆️"
> 		if order == "desc" {
> 			arrow = "⬇️"
> 		}
> 		fmt.Printf("  %s Sort by: %s (%s)\n", arrow, sort, order)
> 	}
> 	
> 	from := params.Get("from")
> 	to := params.Get("to")
> 	if from != "" && to != "" {
> 		fmt.Printf("  📄 Pagination: from %s to %s\n", from, to)
> 	}
> 
> 	if len(params) == 0 {
> 		fmt.Println("  No filters applied (view all)")
> 	}
> 
> 	fmt.Println()
> }
> 
> func viewServers(params url.Values) *ViewServerResponse {
> 	baseURL := "http://localhost:8080/api/v1/servers"
> 	
> 	queryURL := baseURL
> 	if len(params) > 0 {
> 		queryURL += "?" + params.Encode()
> 	}
> 
> 	fmt.Printf("📋 Calling view API: %s\n", queryURL)
> 	fmt.Println()
> 
> 	resp, err := http.Get(queryURL)
> 	if err != nil {
> 		log.Fatalf("Error calling view API: %v", err)
> 	}
> 	defer resp.Body.Close()
> 
> 	if resp.StatusCode != 200 {
> 		body, _ := io.ReadAll(resp.Body)
> 		log.Fatalf("View API failed (Status: %d): %s", resp.StatusCode, string(body))
> 	}
> 
> 	var response ViewServerResponse
> 	body, _ := io.ReadAll(resp.Body)
> 	if err := json.Unmarshal(body, &response); err != nil {
> 		log.Fatalf("Error parsing response: %v", err)
> 	}
> 
> 	return &response
> }
> 
> func printViewResults(response *ViewServerResponse) {
> 	fmt.Println("📊 View Results:")
> 	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════════")
> 
> 	if response.Servers == nil || len(*response.Servers) == 0 {
> 		fmt.Println("❌ No servers found matching the criteria")
> 		return
> 	}
> 
> 	fmt.Printf("📈 Total servers found: %d\n", response.Total)
> 	fmt.Printf("📋 Servers returned: %d\n\n", len(*response.Servers))
> 
> 	// Print header
> 	fmt.Printf("%-12s %-35s %-15s %-8s %-20s %-20s\n",
> 		"ID", "NAME", "IPv4", "STATUS", "CREATED AT", "UPDATED AT")
> 	fmt.Println("─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────")
> 
> 	// Print servers
> 	for i, server := range *response.Servers {
> 		// Truncate long names
> 		name := server.Name
> 		if len(name) > 33 {
> 			name = name[:30] + "..."
> 		}
> 
> 		// Format timestamps
> 		createdAt := server.CreatedAt.Format("2006-01-02 15:04")
> 		updatedAt := server.UpdatedAt.Format("2006-01-02 15:04")
> 
> 		// Status with color indicators
> 		statusIcon := "🔴"
> 		if server.Status == "ON" {
> 			statusIcon = "🟢"
> 		}
> 
> 		fmt.Printf("%-12s %-35s %-15s %s %-6s %-20s %-20s\n",
> 			server.ID,
> 			name,
> 			server.IPv4,
> 			statusIcon,
> 			server.Status,
> 			createdAt,
> 			updatedAt)
> 
> 		// Add separator every 10 rows for better readability
> 		if (i+1)%10 == 0 && i+1 < len(*response.Servers) {
> 			fmt.Println("┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈")
> 		}
> 	}
> 
> 	// Summary statistics
> 	fmt.Println("═══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════")
> 	printSummaryStats(*response.Servers, response.Total)
> }
> 
> func printSummaryStats(servers []Server, total int) {
> 	onCount := 0
> 	offCount := 0
> 	ipPrefixes := make(map[string]int)
> 	nameTypes := make(map[string]int)
> 
> 	for _, server := range servers {
> 		if server.Status == "ON" {
> 			onCount++
> 		} else {
> 			offCount++
> 		}
> 
> 		// Count IP prefixes
> 		if len(server.IPv4) >= 7 {
> 			prefix := server.IPv4[:7] // e.g., "192.168", "10.0.0.", etc.
> 			ipPrefixes[prefix]++
> 		}
> 
> 		// Count server types from names
> 		if len(server.Name) > 0 {
> 			// Extract first word as server type
> 			parts := []string{}
> 			currentWord := ""
> 			for _, char := range server.Name {
> 				if char == ' ' {
> 					if currentWord != "" {
> 						parts = append(parts, currentWord)
> 						break
> 					}
> 				} else {
> 					currentWord += string(char)
> 				}
> 			}
> 			if currentWord != "" && len(parts) == 0 {
> 				parts = append(parts, currentWord)
> 			}
> 			
> 			if len(parts) > 0 {
> 				serverType := parts[0]
> 				nameTypes[serverType]++
> 			}
> 		}
> 	}
> 
> 	fmt.Println("📊 Summary Statistics:")
> 	
> 	if len(servers) > 0 {
> 		fmt.Printf("  🟢 Online servers:  %d (%.1f%%)\n", onCount, float64(onCount)/float64(len(servers))*100)
> 		fmt.Printf("  🔴 Offline servers: %d (%.1f%%)\n", offCount, float64(offCount)/float64(len(servers))*100)
> 	}
> 
> 	if total > len(servers) {
> 		fmt.Printf("  📄 Showing %d of %d total servers\n", len(servers), total)
> 	}
> 
> 	if len(ipPrefixes) > 0 {
> 		fmt.Println("  🌐 Network distribution:")
> 		for prefix, count := range ipPrefixes {
> 			fmt.Printf("     %s*: %d servers\n", prefix, count)
> 		}
> 	}
> 
> 	if len(nameTypes) > 0 {
> 		fmt.Println("  🏷️  Server types:")
> 		for serverType, count := range nameTypes {
> 			fmt.Printf("     %s: %d servers\n", serverType, count)
> 		}
> 	}
> 
> 	fmt.Printf("\n✅ View completed at: %s\n", time.Now().Format("2006-01-02 15:04:05"))
> }
> ````
> 
> ## Ưu điểm của script này:
> 
> ### 1. **Tính năng đa dạng**
> - **Random parameters**: Tự động tạo các tham số filter, sort, pagination ngẫu nhiên
> - **Flexible filtering**: Hỗ trợ filter theo name, status, IPv4
> - **Smart pagination**: Random pagination với size hợp lý (5-25 records)
> - **Multiple sorting**: Sort theo các field khác nhau với order asc/desc
> 
> ### 2. **Hiển thị thông minh**
> - **Color indicators**: 🟢 cho server ON, 🔴 cho server OFF
> - **Table formatting**: Hiển thị dạng bảng rõ ràng với alignment
> - **Text truncation**: Cắt tên server dài để không làm vỡ layout
> - **Pagination separator**: Thêm separator mỗi 10 rows để dễ đọc
> 
> ### 3. **Thống kê chi tiết**
> - **Status distribution**: Phần trăm server online/offline
> - **Network analysis**: Phân tích phân bố IP theo prefix
> - **Server type counting**: Đếm loại server theo tên
> - **Pagination info**: Hiển thị số record trả về vs tổng số
> 
> ### 4. **Error handling**
> - Kiểm tra HTTP status code
> - Parse JSON response an toàn
> - Log lỗi chi tiết khi API call fail
> 
> ### 5. **User experience**
> - **Clear output**: Format output dễ đọc với icons và colors
> - **Progress indication**: Hiển thị URL được gọi
> - **Timestamp**: Ghi nhận thời gian hoàn thành
> 
> Để chạy script:
> 
> ```bash
> cd cmd/scripts/view_server
> go run main.go
> ```
> 
> Script sẽ tự động tạo các tham số ngẫu nhiên và gọi API `ViewServer` để test các tính năng filtering, sorting và pagination của hệ thống.

@workspace gin hoạt động shouldbindquery với server.QueryServerRequest như thế nào

> Copilot: Gin framework hoạt động với `ShouldBindQuery` và `server.QueryServerRequest` thông qua struct tags và reflection. Đây là cách hoạt động chi tiết:
> 
> ## 1. **Struct Tags trong QueryServerRequest**
> 
> Từ constracts.go, bạn có thể thấy struct được định nghĩa với các tags:
> 
> ```go
> type ServerFilter struct {
>     Name   string `json:"name,omitempty" form:"name"`
>     Status string `json:"status,omitempty" form:"status"` 
>     IPv4   string `json:"ipv4,omitempty" form:"ipv4"`
> }
> 
> type ServerSort struct {
>     Sort  string    `json:"sort,omitempty" validate:"omitempty,oneof=name status created_at updated_at" form:"sort"`
>     Order SortOrder `json:"order,omitempty" validate:"omitempty,oneof=asc desc" form:"order"`
> }
> 
> type ServerPagination struct {
>     From int `json:"from,omitempty" form:"from"`
>     To   int `json:"to,omitempty" form:"to"`
> }
> 
> type QueryServerRequest struct {
>     Filter     ServerFilter     `json:"filter"`
>     Pagination ServerPagination `json:"pagination"`
>     Sort       ServerSort       `json:"sort"`
> }
> ```
> 
> ## 2. **Mapping URL Query Parameters**
> 
> Tag `form:"fieldname"` cho Gin biết cách map query parameters từ URL:
> 
> ### Ví dụ URL:
> ```
> GET /api/v1/servers?name=Web&status=ON&ipv4=192.168&sort=name&order=asc&from=0&to=10
> ```
> 
> ### Gin sẽ map như sau:
> - `name=Web` → `req.Filter.Name = "Web"`
> - `status=ON` → `req.Filter.Status = "ON"`  
> - `ipv4=192.168` → `req.Filter.IPv4 = "192.168"`
> - `sort=name` → `req.Sort.Sort = "name"`
> - `order=asc` → `req.Sort.Order = "asc"`
> - `from=0` → `req.Pagination.From = 0`
> - `to=10` → `req.Pagination.To = 10`
> 
> ## 3. **Process trong ViewServer Handler**
> 
> ```go
> func (h *ServerHandler) ViewServer(c *gin.Context) {
>     var req server.QueryServerRequest
>     
>     // Gin tự động parse query parameters và bind vào struct
>     if err := c.ShouldBindQuery(&req); err != nil {
>         h.logger.Error("Failed to bind query", "error", err)
>         c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
>         return
>     }
>     
>     // req đã chứa đầy đủ dữ liệu từ query parameters
>     response, err := h.service.ViewServer(c.Request.Context(), req)
>     // ...
> }
> ```
> 
> ## 4. **Validation và Type Conversion**
> 
> - **String fields**: Gin tự động bind string values
> - **Int fields**: Gin tự động convert string → int
> - **Enum validation**: Tag `validate:"oneof=..."` kiểm tra giá trị hợp lệ
> - **Empty values**: Tag `omitempty` cho phép field rỗng
> 
> ## 5. **Error Handling**
> 
> Nếu query parameters không đúng format:
> ```go
> // URL: /servers?from=abc (from phải là số)
> if err := c.ShouldBindQuery(&req); err != nil {
>     // err sẽ chứa thông tin lỗi conversion
>     return
> }
> ```
> 
> ## 6. **Ưu điểm của cách này**
> 
> ### **Type Safety**: 
> Compile-time checking thay vì runtime parsing
> 
> ### **Automatic Conversion**: 
> Gin tự động convert types (string → int, string → enum)
> 
> ### **Validation**: 
> Dùng với thư viện validator để validate constraints
> 
> ### **Clean Code**: 
> Không cần manual parsing từng query parameter
> 
> ## 7. **Ví dụ Script sử dụng**
> 
> Như trong main.go, script tạo URL với query parameters:
> 
> ```go
> // Generate random parameters
> params := url.Values{}
> params.Add("name", "Web")
> params.Add("status", "ON") 
> params.Add("sort", "name")
> params.Add("order", "asc")
> params.Add("from", "0")
> params.Add("to", "10")
> 
> // Create URL: /servers?name=Web&status=ON&sort=name&order=asc&from=0&to=10
> queryURL := baseURL + "?" + params.Encode()
> ```
> 
> Khi request này đến `ViewServer` handler, Gin sẽ tự động parse và fill vào `QueryServerRequest` struct một cách chính xác và type-safe.

@workspace điều gì khiến gin có thể map chính xác với field như vậy

> Copilot: Gin có thể map chính xác với `server.QueryServerRequest` nhờ vào **reflection** và **struct tags**. Đây là cơ chế hoạt động chi tiết:
> 
> ## 1. **Reflection trong Go**
> 
> Gin sử dụng Go reflection để:
> - **Phân tích cấu trúc struct** tại runtime
> - **Đọc metadata** từ struct tags
> - **Tự động set giá trị** cho các field
> 
> ```go
> // Gin internally sử dụng reflect package
> func (c *Context) ShouldBindQuery(obj interface{}) error {
>     // 1. Lấy reflect.Type và reflect.Value của struct
>     typ := reflect.TypeOf(obj).Elem()
>     val := reflect.ValueOf(obj).Elem()
>     
>     // 2. Iterate qua từng field
>     for i := 0; i < typ.NumField(); i++ {
>         field := typ.Field(i)
>         fieldValue := val.Field(i)
>         
>         // 3. Đọc struct tag để biết tên query parameter
>         tagName := field.Tag.Get("form")
>         
>         // 4. Lấy giá trị từ URL query và set vào field
>         queryValue := c.Query(tagName)
>         // ... convert và set value
>     }
> }
> ```
> 
> ## 2. **Struct Tags - Metadata Key**
> 
> Struct tags là metadata được đính kèm với từng field:
> 
> ```go
> type ServerFilter struct {
>     Name   string `form:"name"`     // "name" parameter → Name field
>     Status string `form:"status"`   // "status" parameter → Status field  
>     IPv4   string `form:"ipv4"`     // "ipv4" parameter → IPv4 field
> }
> 
> type ServerSort struct {
>     Sort  string `form:"sort"`      // "sort" parameter → Sort field
>     Order string `form:"order"`     // "order" parameter → Order field
> }
> ```
> 
> ## 3. **Nested Struct Handling**
> 
> Gin xử lý nested struct bằng cách:
> 
> ```go
> type QueryServerRequest struct {
>     Filter     ServerFilter     `json:"filter"`     // Embedded struct
>     Pagination ServerPagination `json:"pagination"` 
>     Sort       ServerSort       `json:"sort"`
> }
> ```
> 
> Gin sẽ **flatten** các field từ embedded struct:
> - `req.Filter.Name` ← query parameter `name`
> - `req.Filter.Status` ← query parameter `status`
> - `req.Sort.Sort` ← query parameter `sort`
> 
> ## 4. **Type Conversion Magic**
> 
> Gin tự động convert string query parameters thành đúng type:
> 
> ```go
> type ServerPagination struct {
>     From int `form:"from"`  // "5" string → 5 int
>     To   int `form:"to"`    // "10" string → 10 int
> }
> ```
> 
> **Conversion process:**
> ```go
> // URL: ?from=5&to=10
> func convertStringToInt(value string, field reflect.Value) {
>     if intVal, err := strconv.Atoi(value); err == nil {
>         field.SetInt(int64(intVal))
>     }
> }
> ```
> 
> ## 5. **Form Tag Priority**
> 
> Gin theo thứ tự ưu tiên khi đọc tags:
> 
> ```go
> type ServerFilter struct {
>     Name string `form:"name" json:"name"`
>     //           ↑ form tag được ưu tiên cho query binding
> }
> ```
> 
> ## 6. **Error Handling**
> 
> Nếu conversion thất bại:
> 
> ```go
> // URL: ?from=abc (không phải số)
> if err := c.ShouldBindQuery(&req); err != nil {
>     // err sẽ chứa: "strconv.Atoi: parsing \"abc\": invalid syntax"
>     return
> }
> ```
> 
> ## 7. **Complete Flow**
> 
> ```go
> // 1. URL Request
> GET /api/v1/servers?name=Web&status=ON&sort=name&order=asc&from=0&to=10
> 
> // 2. Gin Processing
> func (h *ServerHandler) ViewServer(c *gin.Context) {
>     var req server.QueryServerRequest
>     
>     // 3. Reflection + Tag Reading + Type Conversion
>     if err := c.ShouldBindQuery(&req); err != nil {
>         return
>     }
>     
>     // 4. Result: req được populate đầy đủ
>     // req.Filter.Name = "Web"
>     // req.Filter.Status = "ON"  
>     // req.Sort.Sort = "name"
>     // req.Sort.Order = "asc"
>     // req.Pagination.From = 0
>     // req.Pagination.To = 10
> }
> ```
> 
> ## 8. **Performance Considerations**
> 
> **Reflection có overhead**, nhưng Gin optimize bằng cách:
> - **Cache struct metadata** sau lần đầu analyze
> - **Reuse reflection objects** 
> - **Fast path** cho common types
> 
> ## 9. **Validation Integration**
> 
> Gin có thể kết hợp với validator tags:
> 
> ```go
> type ServerSort struct {
>     Sort  string `form:"sort" validate:"omitempty,oneof=name status created_at"`
>     Order string `form:"order" validate:"omitempty,oneof=asc desc"`
> }
> ```
> 
> Đây là lý do tại sao Gin có thể **"magically"** map URL query parameters vào đúng struct fields một cách chính xác và type-safe, tất cả nhờ vào sức mạnh của **Go reflection** và **struct tags**.

@workspace kết hợp servertypes, enviroments, locations để thành filter names

> Copilot: Tôi sẽ kết hợp `serverTypes`, `environments`, `locations` để tạo `filterNames` đa dạng hơn cho việc test filtering.
> 
> ````go
> // ...existing code...
> 
> var (
> 	serverTypes  = []string{"Web", "Database", "Cache", "Load Balancer", "API", "File", "Mail", "DNS", "Backup", "Monitor"}
> 	environments = []string{"Production", "Staging", "Development", "Testing", "Demo"}
> 	locations    = []string{"US-East", "US-West", "EU-Central", "Asia-Pacific", "Canada", "Australia"}
> )
> 
> var (
> 	// Combine server types, environments, and locations for comprehensive filtering
> 	filterNames    = buildFilterNames()
> 	filterStatuses = []string{"", "ON", "OFF"}
> 	// filterIPs      = []string{"", "192.168", "10.", "172.16"}
> 	sortFields = []string{"", "name", "status", "created_at", "updated_at"}
> 	sortOrders = []string{"", "asc", "desc"}
> )
> 
> // buildFilterNames creates a comprehensive list of filter names from different categories
> func buildFilterNames() []string {
> 	var filters []string
> 	
> 	// Add empty filter for "no filter" case
> 	filters = append(filters, "")
> 	
> 	// Add server types (exact matches)
> 	filters = append(filters, serverTypes...)
> 	
> 	// Add environments for filtering servers by environment
> 	filters = append(filters, environments...)
> 	
> 	// Add locations for filtering servers by location
> 	filters = append(filters, locations...)
> 	
> 	// Add some partial matches for more realistic testing
> 	partialMatches := []string{
> 		"Server",     // Will match any server with "Server" in name
> 		"01",         // Will match servers with "01" in name
> 		"Load",       // Will match "Load Balancer" 
> 		"Data",       // Will match "Database"
> 		"Prod",       // Will match "Production"
> 		"Dev",        // Will match "Development"
> 		"US",         // Will match US locations
> 		"EU",         // Will match EU locations
> 		"Asia",       // Will match Asia-Pacific
> 	}
> 	filters = append(filters, partialMatches...)
> 	
> 	return filters
> }
> 
> // ...existing code...
> ````
> 
> ## Ưu điểm của approach này:
> 
> ### 1. **Comprehensive Testing**
> - **Server Types**: Test filtering theo loại server (Web, Database, API...)
> - **Environments**: Test filtering theo môi trường (Production, Staging...)
> - **Locations**: Test filtering theo vị trí địa lý (US-East, EU-Central...)
> - **Partial Matches**: Test khả năng tìm kiếm partial của database
> 
> ### 2. **Realistic Scenarios**
> Vì server names được tạo theo pattern từ main.go:
> 
> ```go
> // From create_server script
> func generateRandomName() string {
>     serverType := serverTypes[rand.Intn(len(serverTypes))]
>     environment := environments[rand.Intn(len(environments))]
>     location := locations[rand.Intn(len(locations))]
>     number := rand.Intn(99) + 1
> 
>     return fmt.Sprintf("%s Server %02d (%s - %s)", serverType, number, environment, location)
> }
> // Result: "Web Server 15 (Production - US-East)"
> ```
> 
> ### 3. **Filter Testing Coverage**
> 
> **Exact Matches**:
> - `name=Web` → Match "Web Server 01 (Production - US-East)"
> - `name=Production` → Match any server in Production environment
> - `name=US-East` → Match any server in US-East location
> 
> **Partial Matches** (sử dụng `applyFilters` với `ILIKE`):
> - `name=Server` → Match tất cả servers (vì tất cả đều có "Server")
> - `name=Load` → Match "Load Balancer Server 05 (Staging - EU-Central)"
> - `name=Data` → Match "Database Server 23 (Development - Asia-Pacific)"
> 
> ### 4. **Enhanced Testing Output**
> 
> Script sẽ hiển thị filtering results rõ ràng hơn:
> 
> ```
> 🎲 Random View Parameters:
> ─────────────────────────────
>   📛 Filter by name: Production
>   🟢 Filter by status: ON
>   ⬆️ Sort by: name (asc)
>   📄 Pagination: from 2 to 15
> 
> 📋 Calling view API: http://localhost:8080/api/v1/servers?name=Production&status=ON&sort=name&order=asc&from=2&to=15
> 
> 📊 View Results:
> ═══════════════════════════════════════════════════════════════════════════════════════
> 📈 Total servers found: 25
> 📋 Servers returned: 13
> 
> ID           NAME                                IPv4            STATUS   CREATED AT           UPDATED AT          
> ─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
> api-0034     API Server 12 (Production - Canad... 172.16.45.123   🟢 ON     2024-01-15 10:30     2024-01-15 10:30
> db-0089      Database Server 08 (Production - U... 192.168.12.45   🟢 ON     2024-01-15 10:25     2024-01-15 10:25
> ```
> 
> Cách này sẽ tạo ra các test cases đa dạng và realistic hơn, giúp verify functionality của filtering system một cách comprehensive.

hàm buildFilterNames được gọi khi nào, trước hay sau main

> Copilot: Hàm `buildFilterNames()` được gọi **TRƯỚC** hàm `main()` - cụ thể là tại thời điểm **package initialization**.
> 
> ## Thứ tự thực thi:
> 
> ### 1. **Package Level Variable Initialization**
> ```go
> var (
>     filterNames = buildFilterNames() // ← Được gọi ĐẦU TIÊN
>     filterStatuses = []string{"", "ON", "OFF"}
>     sortFields = []string{"", "name", "status", "created_at", "updated_at"}
>     sortOrders = []string{"", "asc", "desc"}
> )
> ```
> 
> ### 2. **Main Function Execution**
> ```go
> func main() {
>     // ← Được gọi SAU khi tất cả package variables đã được khởi tạo
>     rand.Seed(time.Now().UnixNano())
>     // ...
> }
> ```
> 
> ## Timeline chi tiết:
> 
> ```
> 1. Program starts
> 2. Package initialization begins
> 3. buildFilterNames() executes ← Tại đây
> 4. filterNames variable gets the result
> 5. Other package variables initialize
> 6. Package initialization completes
> 7. main() function starts ← Sau đó mới đến đây
> ```
> 
> ## Tại sao như vậy?
> 
> **Go Package Initialization Order**:
> - Tất cả **package-level variables** được khởi tạo trước khi `main()` chạy
> - Các **function calls** trong variable initialization được thực thi ngay lập tức
> - Đảm bảo tất cả dependencies đã sẵn sàng trước khi `main()` bắt đầu
> 
> ## Ví dụ minh họa:
> 
> ````go
> package main
> 
> import "fmt"
> 
> // buildFilterNames được gọi TẠI ĐÂY (trước main)
> func buildFilterNames() []string {
>     fmt.Println("buildFilterNames() is executing") // In ra TRƯỚC main
>     return []string{"", "Web", "Database", "API"}
> }
> 
> var (
>     filterNames = buildFilterNames() // Gọi function ngay lập tức
> )
> 
> func main() {
>     fmt.Println("main() is executing") // In ra SAU buildFilterNames
>     fmt.Println("filterNames:", filterNames)
> }
> ````
> 
> **Output sẽ là:**
> ```
> buildFilterNames() is executing
> main() is executing  
> filterNames: [ Web Database API]
> ```
> 
> Điều này có nghĩa là khi hàm `main()` bắt đầu chạy, biến `filterNames` đã được khởi tạo sẵn với giá trị trả về từ `buildFilterNames()`.

làm sao để linh hoạt giữa việc chuyển đổi buildFilterNames và  buildBlankFilterNames

> Copilot: Tôi sẽ tạo cơ chế linh hoạt để chuyển đổi giữa `buildFilterNames` và `buildBlankFilterNames` bằng command line flags hoặc environment variables.
> 
> ````go
> package main
> 
> import (
> 	"encoding/json"
> 	"flag"
> 	"fmt"
> 	"io"
> 	"log"
> 	"math/rand"
> 	"net/http"
> 	"net/url"
> 	"os"
> 	"strings"
> 	"time"
> 
> 	"github.com/lits-06/vcs-sms/entity"
> 	"github.com/lits-06/vcs-sms/services/server"
> )
> 
> var (
> 	serverTypes  = []string{"Web", "Database", "Cache", "Load Balancer", "API", "File", "Mail", "DNS", "Backup", "Monitor"}
> 	environments = []string{"Production", "Staging", "Development", "Testing", "Demo"}
> 	locations    = []string{"US-East", "US-West", "EU-Central", "Asia-Pacific", "Canada", "Australia"}
> )
> 
> // Command line flags
> var (
> 	filterMode = flag.String("filter", "full", "Filter mode: 'full', 'blank', 'partial', 'env-only', 'type-only'")
> 	runs       = flag.Int("runs", 1, "Number of test runs to execute")
> 	baseURL    = flag.String("url", "http://localhost:8080", "Base API URL")
> 	verbose    = flag.Bool("v", false, "Verbose output")
> )
> 
> var (
> 	filterNames    []string
> 	filterStatuses = []string{"", "ON", "OFF"}
> 	sortFields     = []string{"", "name", "status", "created_at", "updated_at"}
> 	sortOrders     = []string{"", "asc", "desc"}
> )
> 
> func main() {
> 	// Parse command line flags
> 	flag.Parse()
> 
> 	// Initialize filter names based on mode
> 	initializeFilters()
> 
> 	// Seed random
> 	rand.Seed(time.Now().UnixNano())
> 
> 	fmt.Printf("=== Random Server View Tool (%s mode) ===\n", *filterMode)
> 	fmt.Printf("Running %d test(s) against %s\n", *runs, *baseURL)
> 	fmt.Println()
> 
> 	// Run multiple tests if specified
> 	for i := 0; i < *runs; i++ {
> 		if *runs > 1 {
> 			fmt.Printf("🔄 Test Run #%d/%d\n", i+1, *runs)
> 			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
> 		}
> 
> 		runSingleTest()
> 
> 		if i < *runs-1 {
> 			fmt.Println("\n" + strings.Repeat("═", 80) + "\n")
> 			time.Sleep(500 * time.Millisecond) // Small delay between runs
> 		}
> 	}
> 
> 	if *runs > 1 {
> 		fmt.Printf("\n✅ Completed %d test runs\n", *runs)
> 	}
> }
> 
> func initializeFilters() {
> 	// Check environment variable first
> 	if envMode := os.Getenv("FILTER_MODE"); envMode != "" {
> 		*filterMode = envMode
> 	}
> 
> 	if *verbose {
> 		fmt.Printf("🔧 Initializing filters in '%s' mode\n", *filterMode)
> 	}
> 
> 	switch strings.ToLower(*filterMode) {
> 	case "full", "complete":
> 		filterNames = buildFullFilterNames()
> 		if *verbose {
> 			fmt.Printf("   📝 Full filters: %d options available\n", len(filterNames))
> 		}
> 	case "blank", "empty", "none":
> 		filterNames = buildBlankFilterNames()
> 		if *verbose {
> 			fmt.Printf("   📝 Blank filters: only empty filter available\n")
> 		}
> 	case "partial", "part":
> 		filterNames = buildPartialFilterNames()
> 		if *verbose {
> 			fmt.Printf("   📝 Partial filters: %d partial match options\n", len(filterNames))
> 		}
> 	case "env-only", "environment":
> 		filterNames = buildEnvironmentOnlyFilters()
> 		if *verbose {
> 			fmt.Printf("   📝 Environment filters: %d environments\n", len(filterNames))
> 		}
> 	case "type-only", "types":
> 		filterNames = buildTypeOnlyFilters()
> 		if *verbose {
> 			fmt.Printf("   📝 Type filters: %d server types\n", len(filterNames))
> 		}
> 	case "location-only", "locations":
> 		filterNames = buildLocationOnlyFilters()
> 		if *verbose {
> 			fmt.Printf("   📝 Location filters: %d locations\n", len(filterNames))
> 		}
> 	default:
> 		log.Fatalf("Invalid filter mode: %s. Valid options: full, blank, partial, env-only, type-only, location-only", *filterMode)
> 	}
> 
> 	if *verbose {
> 		fmt.Println()
> 	}
> }
> 
> // Different filter building strategies
> func buildFullFilterNames() []string {
> 	var filters []string
> 	filters = append(filters, "") // Empty filter
> 	filters = append(filters, serverTypes...)
> 	filters = append(filters, environments...)
> 	filters = append(filters, locations...)
> 	
> 	// Add partial matches
> 	partialMatches := []string{"Server", "01", "Load", "Data", "Prod", "Dev", "US", "EU", "Asia"}
> 	filters = append(filters, partialMatches...)
> 	
> 	return filters
> }
> 
> func buildBlankFilterNames() []string {
> 	return []string{""}
> }
> 
> func buildPartialFilterNames() []string {
> 	return []string{"", "Server", "01", "02", "Load", "Data", "Prod", "Dev", "Test", "US", "EU", "Asia"}
> }
> 
> func buildEnvironmentOnlyFilters() []string {
> 	filters := []string{""}
> 	filters = append(filters, environments...)
> 	return filters
> }
> 
> func buildTypeOnlyFilters() []string {
> 	filters := []string{""}
> 	filters = append(filters, serverTypes...)
> 	return filters
> }
> 
> func buildLocationOnlyFilters() []string {
> 	filters := []string{""}
> 	filters = append(filters, locations...)
> 	return filters
> }
> 
> func runSingleTest() {
> 	// Generate random parameters
> 	params := generateRandomParams()
> 
> 	// Print parameters being used
> 	printViewParams(params)
> 
> 	// Call view API and get results
> 	response := viewServers(params)
> 
> 	// Print results
> 	printViewResults(response)
> }
> 
> // ...existing generateRandomParams function...
> 
> func printViewParams(params url.Values) {
> 	fmt.Printf("🎲 Random View Parameters (%s mode):\n", *filterMode)
> 	fmt.Println("─────────────────────────────")
> 
> 	if name := params.Get("name"); name != "" {
> 		filterType := determineFilterType(name)
> 		fmt.Printf("  📛 Filter by name: %s %s\n", name, filterType)
> 	}
> 
> 	if status := params.Get("status"); status != "" {
> 		statusIcon := "🔴"
> 		if status == "ON" {
> 			statusIcon = "🟢"
> 		}
> 		fmt.Printf("  %s Filter by status: %s\n", statusIcon, status)
> 	}
> 
> 	if ipv4 := params.Get("ipv4"); ipv4 != "" {
> 		fmt.Printf("  🌐 Filter by IPv4: %s*\n", ipv4)
> 	}
> 
> 	if sort := params.Get("sort"); sort != "" {
> 		order := params.Get("order")
> 		if order == "" {
> 			order = "asc"
> 		}
> 		arrow := "⬆️"
> 		if order == "desc" {
> 			arrow = "⬇️"
> 		}
> 		fmt.Printf("  %s Sort by: %s (%s)\n", arrow, sort, order)
> 	}
> 
> 	from := params.Get("from")
> 	to := params.Get("to")
> 	if from != "" && to != "" {
> 		fmt.Printf("  📄 Pagination: from %s to %s\n", from, to)
> 	}
> 
> 	if len(params) == 2 && params.Get("from") != "" && params.Get("to") != "" {
> 		fmt.Printf("  ℹ️  No filters applied (view all with pagination)\n")
> 	}
> 
> 	fmt.Println()
> }
> 
> func determineFilterType(name string) string {
> 	for _, t := range serverTypes {
> 		if t == name {
> 			return "(server type)"
> 		}
> 	}
> 	for _, e := range environments {
> 		if e == name {
> 			return "(environment)"
> 		}
> 	}
> 	for _, l := range locations {
> 		if l == name {
> 			return "(location)"
> 		}
> 	}
> 	return "(partial match)"
> }
> 
> func viewServers(params url.Values) *server.QueryServerResponse {
> 	queryURL := *baseURL + "/api/v1/servers"
> 	if len(params) > 0 {
> 		queryURL += "?" + params.Encode()
> 	}
> 
> 	if *verbose {
> 		fmt.Printf("📋 Calling view API: %s\n", queryURL)
> 	} else {
> 		fmt.Printf("📋 Calling API with %d parameter(s)\n", len(params))
> 	}
> 	fmt.Println()
> 
> 	resp, err := http.Get(queryURL)
> 	if err != nil {
> 		log.Fatalf("Error calling view API: %v", err)
> 	}
> 	defer resp.Body.Close()
> 
> 	if resp.StatusCode != 200 {
> 		body, _ := io.ReadAll(resp.Body)
> 		log.Fatalf("View API failed (Status: %d): %s", resp.StatusCode, string(body))
> 	}
> 
> 	var response server.QueryServerResponse
> 	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
> 		log.Fatalf("Error decoding response: %v", err)
> 	}
> 
> 	return &response
> }
> 
> // ...existing generateRandomParams, printViewResults, printSummaryStats functions...
> ````
> 
> ## Cách sử dụng linh hoạt:
> 
> ### 1. **Command Line Flags**
> ```bash
> # Full filtering (default)
> go run main.go
> 
> # Blank filtering only
> go run main.go -filter=blank
> 
> # Environment filtering only  
> go run main.go -filter=env-only
> 
> # Server type filtering only
> go run main.go -filter=type-only
> 
> # Location filtering only
> go run main.go -filter=location-only
> 
> # Partial matching only
> go run main.go -filter=partial
> 
> # Multiple runs with verbose output
> go run main.go -filter=full -runs=5 -v
> 
> # Custom API URL
> go run main.go -url=http://prod-server:8080 -filter=env-only
> ```
> 
> ### 2. **Environment Variables**
> ```bash
> # Set filter mode via environment
> export FILTER_MODE=blank
> go run main.go
> 
> # Or inline
> FILTER_MODE=env-only go run main.go
> ```
> 
> ### 3. **Makefile Integration**
> ```makefile
> # Add to Makefile
> test-view-full:
> 	cd cmd/scripts/view_server && go run main.go -filter=full -runs=3
> 
> test-view-blank:
> 	cd cmd/scripts/view_server && go run main.go -filter=blank -runs=5
> 
> test-view-env:
> 	cd cmd/scripts/view_server && go run main.go -filter=env-only -runs=2 -v
> ```
> 
> ## Ưu điểm của approach này:
> 
> ### 1. **Flexible Testing**
> - **Development**: Dùng `blank` mode để test without filters
> - **Feature Testing**: Dùng `env-only` để test environment filtering  
> - **Full Integration**: Dùng `full` mode cho comprehensive testing
> 
> ### 2. **CI/CD Integration**
> ```yaml
> # GitHub Actions example
> - name: Test View API - No Filters
>   run: FILTER_MODE=blank go run cmd/scripts/view_server/main.go -runs=3
> 
> - name: Test View API - Full Filters  
>   run: FILTER_MODE=full go run cmd/scripts/view_server/main.go -runs=5
> ```
> 
> ### 3. **Progressive Testing**
> - Start với `blank` mode để đảm bảo basic functionality
> - Move to `partial` mode để test search capabilities
> - Use `full` mode để test comprehensive filtering
> 
> ### 4. **Debug Capabilities**
> - Use `-v` flag để see detailed request URLs
> - Use `-runs=N` để test consistency across multiple runs
> - Different modes help isolate specific filtering issues
> 
> Cách này cho phép bạn linh hoạt switch giữa các testing strategies mà không cần modify code, rất hữu ích cho development và testing workflows khác nhau.

có thể làm thêm hướng dẫn về các flag không

> Copilot: Tôi sẽ thêm hướng dẫn chi tiết về các command line flags cho script view server.
> 
> ````go
> package main
> 
> import (
> 	"encoding/json"
> 	"flag"
> 	"fmt"
> 	"io"
> 	"log"
> 	"math/rand"
> 	"net/http"
> 	"net/url"
> 	"os"
> 	"strings"
> 	"time"
> 
> 	"github.com/lits-06/vcs-sms/entity"
> 	"github.com/lits-06/vcs-sms/services/server"
> )
> 
> var (
> 	serverTypes  = []string{"Web", "Database", "Cache", "Load Balancer", "API", "File", "Mail", "DNS", "Backup", "Monitor"}
> 	environments = []string{"Production", "Staging", "Development", "Testing", "Demo"}
> 	locations    = []string{"US-East", "US-West", "EU-Central", "Asia-Pacific", "Canada", "Australia"}
> )
> 
> // Command line flags with detailed descriptions
> var (
> 	filterMode = flag.String("filter", "full", 
> 		"Filter mode strategy:\n"+
> 		"  • full      - All filters (types, environments, locations, partials)\n"+
> 		"  • blank     - No name filters (only status/pagination)\n"+
> 		"  • partial   - Partial string matches only\n"+
> 		"  • env-only  - Environment filters only\n"+
> 		"  • type-only - Server type filters only\n"+
> 		"  • location-only - Location filters only")
> 	
> 	runs = flag.Int("runs", 1, 
> 		"Number of test runs to execute (1-100)")
> 	
> 	baseURL = flag.String("url", "http://localhost:8080", 
> 		"Base API URL (without /api/v1/servers path)")
> 	
> 	verbose = flag.Bool("v", false, 
> 		"Verbose output (show detailed request info)")
> 	
> 	help = flag.Bool("help", false, 
> 		"Show detailed help and usage examples")
> 	
> 	seed = flag.Int64("seed", 0, 
> 		"Random seed for reproducible results (0 = use current time)")
> 	
> 	showStats = flag.Bool("stats", true, 
> 		"Show summary statistics after results")
> )
> 
> var (
> 	filterNames    []string
> 	filterStatuses = []string{"", "ON", "OFF"}
> 	sortFields     = []string{"", "name", "status", "created_at", "updated_at"}
> 	sortOrders     = []string{"", "asc", "desc"}
> )
> 
> func main() {
> 	// Custom usage function
> 	flag.Usage = showUsage
> 	flag.Parse()
> 
> 	// Show help if requested
> 	if *help {
> 		showDetailedHelp()
> 		return
> 	}
> 
> 	// Validate inputs
> 	if err := validateFlags(); err != nil {
> 		fmt.Fprintf(os.Stderr, "❌ Error: %v\n\n", err)
> 		flag.Usage()
> 		os.Exit(1)
> 	}
> 
> 	// Initialize random seed
> 	initializeRandom()
> 
> 	// Initialize filter names based on mode
> 	initializeFilters()
> 
> 	fmt.Printf("=== Random Server View Tool (%s mode) ===\n", *filterMode)
> 	if *runs > 1 {
> 		fmt.Printf("Running %d test(s) against %s\n", *runs, *baseURL)
> 	}
> 	fmt.Println()
> 
> 	// Run tests
> 	runTests()
> }
> 
> func showUsage() {
> 	fmt.Fprintf(os.Stderr, `
> 🔍 Server View Testing Tool
> 
> USAGE:
>   %s [FLAGS]
> 
> FLAGS:
> `, os.Args[0])
> 	flag.PrintDefaults()
> 	
> 	fmt.Fprintf(os.Stderr, `
> EXAMPLES:
>   %s                           # Default: full filtering, single run
>   %s -filter=blank -runs=5     # Test without name filters, 5 runs
>   %s -filter=env-only -v       # Environment filtering with verbose output
>   %s -url=http://prod:8080     # Test against production server
>   %s -help                     # Show detailed help
> 
> ENVIRONMENT VARIABLES:
>   FILTER_MODE    Override filter mode (same values as -filter flag)
> 
> `, os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0])
> }
> 
> func showDetailedHelp() {
> 	fmt.Print(`
> 🔍 Server View Testing Tool - Detailed Help
> ═══════════════════════════════════════════
> 
> OVERVIEW:
>   This tool tests the server view API by generating random filter, sort,
>   and pagination parameters. It helps verify API functionality across
>   different scenarios.
> 
> FILTER MODES:
>   ┌─────────────┬──────────────────────────────────────────────────────┐
>   │ Mode        │ Description                                          │
>   ├─────────────┼──────────────────────────────────────────────────────┤
>   │ full        │ All available filters: server types, environments,  │
>   │             │ locations, and partial string matches               │
>   │             │ Examples: "Web", "Production", "US-East", "Server"  │
>   ├─────────────┼──────────────────────────────────────────────────────┤
>   │ blank       │ No name filtering, only status and pagination       │
>   │             │ Useful for testing base functionality               │
>   ├─────────────┼──────────────────────────────────────────────────────┤
>   │ partial     │ Partial string matches only                         │
>   │             │ Examples: "Server", "01", "Load", "Data"            │
>   ├─────────────┼──────────────────────────────────────────────────────┤
>   │ env-only    │ Environment filters only                            │
>   │             │ Examples: "Production", "Staging", "Development"    │
>   ├─────────────┼──────────────────────────────────────────────────────┤
>   │ type-only   │ Server type filters only                            │
>   │             │ Examples: "Web", "Database", "API", "Cache"         │
>   ├─────────────┼──────────────────────────────────────────────────────┤
>   │ location-only│ Location filters only                              │
>   │             │ Examples: "US-East", "EU-Central", "Asia-Pacific"   │
>   └─────────────┴──────────────────────────────────────────────────────┘
> 
> USAGE EXAMPLES:
> 
>   🚀 Quick Testing:
>     go run main.go                    # Single test with full filtering
>     go run main.go -filter=blank      # Test without name filters
> 
>   📊 Multiple Runs:
>     go run main.go -runs=10           # Run 10 tests for consistency
>     go run main.go -runs=5 -v         # 5 runs with verbose output
> 
>   🎯 Specific Filter Testing:
>     go run main.go -filter=env-only -runs=3    # Test environment filtering
>     go run main.go -filter=type-only -v        # Test server type filtering
> 
>   🌐 Different Environments:
>     go run main.go -url=http://staging:8080    # Test staging environment
>     go run main.go -url=https://api.prod.com   # Test production API
> 
>   🔍 Debugging:
>     go run main.go -v -seed=12345              # Verbose with fixed seed
>     go run main.go -filter=partial -runs=1 -v # Debug partial filtering
> 
>   📈 Performance Testing:
>     go run main.go -runs=50 -stats=false      # Many runs, no detailed stats
> 
> ENVIRONMENT VARIABLES:
>   FILTER_MODE=env-only go run main.go         # Override filter mode
>   
> RANDOM PARAMETERS GENERATED:
>   • Name filters (based on selected mode)
>   • Status filters (ON, OFF, or empty)
>   • Sort fields (name, status, created_at, updated_at)
>   • Sort orders (asc, desc)
>   • Pagination (random from/to ranges)
> 
> OUTPUT EXPLANATION:
>   🎲 Parameters section shows what filters were randomly selected
>   📋 API call section shows the actual URL being called
>   📊 Results section shows servers found and detailed statistics
>   📈 Summary section shows distribution analysis
> 
> TIPS:
>   • Use -v flag to see actual API URLs for debugging
>   • Use -seed flag to reproduce specific test scenarios
>   • Use different -runs values to test consistency
>   • Combine with shell scripts for automated testing
> 
> RETURN CODES:
>   0  Success
>   1  Invalid arguments or API errors
> `)
> }
> 
> func validateFlags() error {
> 	// Validate runs
> 	if *runs < 1 || *runs > 100 {
> 		return fmt.Errorf("runs must be between 1 and 100, got %d", *runs)
> 	}
> 
> 	// Validate filter mode
> 	validModes := []string{"full", "blank", "partial", "env-only", "type-only", "location-only", "complete", "empty", "none", "environment", "types", "locations", "part"}
> 	valid := false
> 	for _, mode := range validModes {
> 		if strings.ToLower(*filterMode) == mode {
> 			valid = true
> 			break
> 		}
> 	}
> 	if !valid {
> 		return fmt.Errorf("invalid filter mode '%s'. Valid options: full, blank, partial, env-only, type-only, location-only", *filterMode)
> 	}
> 
> 	// Validate URL format
> 	if !strings.HasPrefix(*baseURL, "http://") && !strings.HasPrefix(*baseURL, "https://") {
> 		return fmt.Errorf("base URL must start with http:// or https://, got '%s'", *baseURL)
> 	}
> 
> 	return nil
> }
> 
> func initializeRandom() {
> 	if *seed != 0 {
> 		rand.Seed(*seed)
> 		if *verbose {
> 			fmt.Printf("🎲 Using fixed random seed: %d (for reproducible results)\n", *seed)
> 		}
> 	} else {
> 		rand.Seed(time.Now().UnixNano())
> 	}
> }
> 
> func runTests() {
> 	successCount := 0
> 	
> 	for i := 0; i < *runs; i++ {
> 		if *runs > 1 {
> 			fmt.Printf("🔄 Test Run #%d/%d\n", i+1, *runs)
> 			if i > 0 {
> 				fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
> 			}
> 		}
> 
> 		success := runSingleTest()
> 		if success {
> 			successCount++
> 		}
> 
> 		if i < *runs-1 {
> 			fmt.Println("\n" + strings.Repeat("═", 80) + "\n")
> 			time.Sleep(300 * time.Millisecond) // Small delay between runs
> 		}
> 	}
> 
> 	// Final summary for multiple runs
> 	if *runs > 1 {
> 		fmt.Printf("\n🏁 Test Summary: %d/%d successful runs", successCount, *runs)
> 		if successCount == *runs {
> 			fmt.Println(" ✅")
> 		} else {
> 			fmt.Printf(" ⚠️  (%d failed)\n", *runs-successCount)
> 		}
> 	}
> }
> 
> func runSingleTest() bool {
> 	defer func() {
> 		if r := recover(); r != nil {
> 			fmt.Printf("❌ Test failed with panic: %v\n", r)
> 		}
> 	}()
> 
> 	// Generate random parameters
> 	params := generateRandomParams()
> 
> 	// Print parameters being used
> 	printViewParams(params)
> 
> 	// Call view API and get results
> 	response := viewServers(params)
> 	if response == nil {
> 		return false
> 	}
> 
> 	// Print results
> 	printViewResults(response)
> 	
> 	return true
> }
> 
> func initializeFilters() {
> 	// Check environment variable first
> 	if envMode := os.Getenv("FILTER_MODE"); envMode != "" {
> 		*filterMode = envMode
> 		if *verbose {
> 			fmt.Printf("🌍 Filter mode overridden by environment variable: %s\n", envMode)
> 		}
> 	}
> 
> 	if *verbose {
> 		fmt.Printf("🔧 Initializing filters in '%s' mode\n", *filterMode)
> 	}
> 
> 	switch strings.ToLower(*filterMode) {
> 	case "full", "complete":
> 		filterNames = buildFullFilterNames()
> 		if *verbose {
> 			fmt.Printf("   📝 Full filters: %d options available\n", len(filterNames))
> 		}
> 	case "blank", "empty", "none":
> 		filterNames = buildBlankFilterNames()
> 		if *verbose {
> 			fmt.Printf("   📝 Blank filters: only empty filter available\n")
> 		}
> 	case "partial", "part":
> 		filterNames = buildPartialFilterNames()
> 		if *verbose {
> 			fmt.Printf("   📝 Partial filters: %d partial match options\n", len(filterNames))
> 		}
> 	case "env-only", "environment":
> 		filterNames = buildEnvironmentOnlyFilters()
> 		if *verbose {
> 			fmt.Printf("   📝 Environment filters: %d environments\n", len(filterNames))
> 		}
> 	case "type-only", "types":
> 		filterNames = buildTypeOnlyFilters()
> 		if *verbose {
> 			fmt.Printf("   📝 Type filters: %d server types\n", len(filterNames))
> 		}
> 	case "location-only", "locations":
> 		filterNames = buildLocationOnlyFilters()
> 		if *verbose {
> 			fmt.Printf("   📝 Location filters: %d locations\n", len(filterNames))
> 		}
> 	}
> 
> 	if *verbose && len(filterNames) > 0 {
> 		fmt.Printf("   🎯 Sample filters: %v\n", filterNames[:min(5, len(filterNames))])
> 		fmt.Println()
> 	}
> }
> 
> // Filter building functions
> func buildFullFilterNames() []string {
> 	var filters []string
> 	filters = append(filters, "") // Empty filter
> 	filters = append(filters, serverTypes...)
> 	filters = append(filters, environments...)
> 	filters = append(filters, locations...)
> 	
> 	// Add partial matches
> 	partialMatches := []string{"Server", "01", "02", "Load", "Data", "Prod", "Dev", "Test", "US", "EU", "Asia"}
> 	filters = append(filters, partialMatches...)
> 	
> 	return filters
> }
> 
> func buildBlankFilterNames() []string {
> 	return []string{""}
> }
> 
> func buildPartialFilterNames() []string {
> 	return []string{"", "Server", "01", "02", "03", "Load", "Data", "Prod", "Dev", "Test", "US", "EU", "Asia", "Cache", "API"}
> }
> 
> func buildEnvironmentOnlyFilters() []string {
> 	filters := []string{""}
> 	filters = append(filters, environments...)
> 	return filters
> }
> 
> func buildTypeOnlyFilters() []string {
> 	filters := []string{""}
> 	filters = append(filters, serverTypes...)
> 	return filters
> }
> 
> func buildLocationOnlyFilters() []string {
> 	filters := []string{""}
> 	filters = append(filters, locations...)
> 	return filters
> }
> 
> func min(a, b int) int {
> 	if a < b {
> 		return a
> 	}
> 	return b
> }
> 
> // ...existing generateRandomParams, printViewParams, viewServers, printViewResults, printSummaryStats functions...
> ````
> 
> ## Cách sử dụng với các flags:
> 
> ### **Quick Start:**
> ```bash
> # Xem help cơ bản
> go run main.go -help
> 
> # Test đơn giản
> go run main.go
> 
> # Test với verbose
> go run main.go -v
> ```
> 
> ### **Filter Mode Testing:**
> ```bash
> # Test all filters
> go run main.go -filter=full -runs=3
> 
> # Test without name filters  
> go run main.go -filter=blank -runs=5
> 
> # Test only environment filters
> go run main.go -filter=env-only -v
> 
> # Test only server type filters
> go run main.go -filter=type-only -runs=2
> ```
> 
> ### **Reproducible Testing:**
> ```bash
> # Fixed seed for debugging
> go run main.go -seed=12345 -v
> 
> # Same test twice
> go run main.go -seed=98765 -runs=2
> ```
> 
> ### **Different Environments:**
> ```bash
> # Test staging
> go run main.go -url=http://staging:8080 -filter=env-only
> 
> # Test production
> go run main.go -url=https://api.production.com -runs=10
> ```
> 
> ### **Performance Testing:**
> ```bash
> # Many runs without detailed stats
> go run main.go -runs=50 -stats=false
> 
> # Quick stress test
> go run main.go -filter=blank -runs=20
> ```
> 
> Với hướng dẫn chi tiết này, users có thể dễ dàng hiểu và sử dụng script để test API view server một cách hiệu quả!

muốn chỉ lấy 1 trạng thái ON hoặc OFF

> Copilot: Tôi sẽ thay đổi code để chỉ lấy 1 trạng thái ON hoặc OFF (loại bỏ trường hợp empty status):
> 
> ````go
> var (
> 	filterNames    = []string{}
> 	filterStatuses = []string{"ON", "OFF"} // Chỉ lấy ON hoặc OFF, bỏ empty string
> 	// filterIPs      = []string{"", "192.168", "10.", "172.16"}
> 	sortFields = []string{"", "name", "status", "created_at", "updated_at"}
> 	sortOrders = []string{"", "asc", "desc"}
> )
> 
> // ...existing code...
> 
> func generateRandomParams() url.Values {
> 	params := url.Values{}
> 
> 	// Random filter by name
> 	filterName := filterNames[rand.Intn(len(filterNames))]
> 	if filterName != "" {
> 		params.Add("name", filterName)
> 	}
> 
> 	// Random filter by status - LUÔN có status (ON hoặc OFF)
> 	filterStatus := filterStatuses[rand.Intn(len(filterStatuses))]
> 	params.Add("status", filterStatus) // Bỏ điều kiện if != ""
> 
> 	// Random filter by IPv4
> 	// filterIP := filterIPs[rand.Intn(len(filterIPs))]
> 	// if filterIP != "" {
> 	// 	params.Add("ipv4", filterIP)
> 	// }
> 
> 	// Random sort
> 	sortField := sortFields[rand.Intn(len(sortFields))]
> 	sortOrder := sortOrders[rand.Intn(len(sortOrders))]
> 	if sortField != "" {
> 		params.Add("sort", sortField)
> 	}
> 	if sortOrder != "" {
> 		params.Add("order", sortOrder)
> 	}
> 
> 	// Random pagination
> 	from := rand.Intn(5)           // Start from 0-4
> 	to := from + rand.Intn(20) + 5 // Get 5-25 records
> 	params.Add("from", fmt.Sprintf("%d", from))
> 	params.Add("to", fmt.Sprintf("%d", to))
> 
> 	return params
> }
> 
> func printViewParams(params url.Values) {
> 	fmt.Println("🎲 Random View Parameters:")
> 	fmt.Println("─────────────────────────────")
> 
> 	if name := params.Get("name"); name != "" {
> 		fmt.Printf("  📛 Filter by name: %s\n", name)
> 	}
> 
> 	// Status sẽ LUÔN có giá trị
> 	if status := params.Get("status"); status != "" {
> 		statusIcon := "🔴"
> 		if status == "ON" {
> 			statusIcon = "🟢"
> 		}
> 		fmt.Printf("  %s Filter by status: %s\n", statusIcon, status)
> 	}
> 
> 	if ipv4 := params.Get("ipv4"); ipv4 != "" {
> 		fmt.Printf("  🌐 Filter by IPv4: %s*\n", ipv4)
> 	}
> 
> 	if sort := params.Get("sort"); sort != "" {
> 		order := params.Get("order")
> 		if order == "" {
> 			order = "asc"
> 		}
> 		arrow := "⬆️"
> 		if order == "desc" {
> 			arrow = "⬇️"
> 		}
> 		fmt.Printf("  %s Sort by: %s (%s)\n", arrow, sort, order)
> 	}
> 
> 	from := params.Get("from")
> 	to := params.Get("to")
> 	if from != "" && to != "" {
> 		fmt.Printf("  📄 Pagination: from %s to %s\n", from, to)
> 	}
> 
> 	// Cập nhật điều kiện kiểm tra "no filters"
> 	nameEmpty := params.Get("name") == ""
> 	if nameEmpty && len(params) == 3 { // chỉ có status, from, to
> 		fmt.Printf("  ℹ️  Only status filter applied (no name filter)\n")
> 	}
> 
> 	fmt.Println()
> }
> ````
> 
> ## Thay đổi chính:
> 
> ### 1. **Loại bỏ empty status**
> ```go
> // Trước:
> filterStatuses = []string{"", "ON", "OFF"}
> 
> // Sau:
> filterStatuses = []string{"ON", "OFF"}
> ```
> 
> ### 2. **Luôn áp dụng status filter**
> ```go
> // Trước:
> filterStatus := filterStatuses[rand.Intn(len(filterStatuses))]
> if filterStatus != "" {
>     params.Add("status", filterStatus)
> }
> 
> // Sau:
> filterStatus := filterStatuses[rand.Intn(len(filterStatuses))]
> params.Add("status", filterStatus) // Luôn thêm status
> ```
> 
> ### 3. **Cập nhật logic hiển thị**
> ```go
> // Cập nhật điều kiện "no filters applied"
> nameEmpty := params.Get("name") == ""
> if nameEmpty && len(params) == 3 { // chỉ có status, from, to
>     fmt.Printf("  ℹ️  Only status filter applied (no name filter)\n")
> }
> ```
> 
> ## Kết quả mong đợi:
> 
> ### **Sample Output:**
> ```
> 🎲 Random View Parameters:
> ─────────────────────────────
>   📛 Filter by name: Database
>   🟢 Filter by status: ON
>   ⬆️ Sort by: name (asc)
>   📄 Pagination: from 2 to 15
> ```
> 
> Hoặc:
> ```
> 🎲 Random View Parameters:
> ─────────────────────────────
>   🔴 Filter by status: OFF
>   ⬇️ Sort by: created_at (desc)
>   📄 Pagination: from 0 to 8
>   ℹ️  Only status filter applied (no name filter)
> ```
> 
> ## Lý do thay đổi:
> 
> ### 1. **Focused Testing**: 
> Luôn test với status filter để verify filtering functionality
> 
> ### 2. **More Realistic**: 
> Trong thực tế, việc filter theo status (ON/OFF) rất common
> 
> ### 3. **Better Coverage**: 
> Đảm bảo cả 2 trạng thái ON và OFF đều được test đều đặn
> 
> ### 4. **Clear Results**: 
> Kết quả sẽ rõ ràng hơn vì luôn có status filter để phân biệt
> 
> Bây giờ mỗi lần chạy script sẽ **luôn** có status filter (ON hoặc OFF), giúp test API filtering functionality một cách consistent hơn.

chỉ định ở terminal

> Copilot: Bạn có thể chỉ định status ON hoặc OFF trực tiếp từ terminal bằng nhiều cách:
> 
> ## 1. **Thêm Flag cho Status**
> 
> ````go
> // Command line flags with detailed descriptions
> var (
> 	filterMode = flag.String("filter", "full",
> 		"Filter mode strategy:\n"+
> 		"  • full      - All filters (types, environments, locations, partials)\n"+
> 		"  • blank     - No name filters (only status/pagination)\n"+
> 		"  • partial   - Partial string matches only\n"+
> 		"  • env-only  - Environment filters only\n"+
> 		"  • type-only - Server type filters only\n"+
> 		"  • location-only - Location filters only")
> 
> 	status = flag.String("status", "", 
> 		"Force specific status filter (ON, OFF, or empty for random)")
> 
> 	verbose = flag.Bool("v", false,
> 		"Verbose output (show detailed request info)")
> 
> 	help = flag.Bool("help", false,
> 		"Show detailed help and usage examples")
> )
> 
> // ...existing code...
> 
> func generateRandomParams() url.Values {
> 	params := url.Values{}
> 
> 	// Random filter by name
> 	filterName := filterNames[rand.Intn(len(filterNames))]
> 	if filterName != "" {
> 		params.Add("name", filterName)
> 	}
> 
> 	// Status filter - check flag first, then random
> 	var filterStatus string
> 	if *status != "" {
> 		// Use specified status from command line
> 		filterStatus = *status
> 		if *verbose {
> 			fmt.Printf("🎯 Using specified status: %s\n", filterStatus)
> 		}
> 	} else {
> 		// Random status
> 		filterStatus = filterStatuses[rand.Intn(len(filterStatuses))]
> 	}
> 	
> 	if filterStatus != "" {
> 		params.Add("status", filterStatus)
> 	}
> 
> 	// Random sort
> 	sortField := sortFields[rand.Intn(len(sortFields))]
> 	sortOrder := sortOrders[rand.Intn(len(sortOrders))]
> 	if sortField != "" {
> 		params.Add("sort", sortField)
> 	}
> 	if sortOrder != "" {
> 		params.Add("order", sortOrder)
> 	}
> 
> 	// Random pagination
> 	from := rand.Intn(5)           // Start from 0-4
> 	to := from + rand.Intn(20) + 5 // Get 5-25 records
> 	params.Add("from", fmt.Sprintf("%d", from))
> 	params.Add("to", fmt.Sprintf("%d", to))
> 
> 	return params
> }
> ````
> 
> ## 2. **Sử dụng từ Terminal**
> 
> ### **Chỉ định status ON:**
> ```bash
> go run main.go -status=ON
> go run main.go -status=ON -filter=full -v
> go run main.go -status=ON -filter=env-only
> ```
> 
> ### **Chỉ định status OFF:**
> ```bash
> go run main.go -status=OFF
> go run main.go -status=OFF -filter=type-only -v
> go run main.go -status=OFF -filter=blank
> ```
> 
> ### **Random status (default):**
> ```bash
> go run main.go
> go run main.go -filter=full
> ```
> 
> ## 3. **Environment Variable**
> 
> ````go
> func generateRandomParams() url.Values {
> 	params := url.Values{}
> 
> 	// Random filter by name
> 	filterName := filterNames[rand.Intn(len(filterNames))]
> 	if filterName != "" {
> 		params.Add("name", filterName)
> 	}
> 
> 	// Status filter - priority: flag > env var > random
> 	var filterStatus string
> 	if *status != "" {
> 		// 1. Command line flag has highest priority
> 		filterStatus = *status
> 		if *verbose {
> 			fmt.Printf("🎯 Using flag status: %s\n", filterStatus)
> 		}
> 	} else if envStatus := os.Getenv("FORCE_STATUS"); envStatus != "" {
> 		// 2. Environment variable
> 		filterStatus = envStatus
> 		if *verbose {
> 			fmt.Printf("🌍 Using env status: %s\n", filterStatus)
> 		}
> 	} else {
> 		// 3. Random (default)
> 		filterStatus = filterStatuses[rand.Intn(len(filterStatuses))]
> 		if *verbose {
> 			fmt.Printf("🎲 Using random status: %s\n", filterStatus)
> 		}
> 	}
> 	
> 	if filterStatus != "" {
> 		params.Add("status", filterStatus)
> 	}
> 
> 	// ...rest of the function...
> }
> ````
> 
> ### **Sử dụng environment variable:**
> ```bash
> # Set environment variable
> export FORCE_STATUS=ON
> go run main.go
> 
> # Or inline
> FORCE_STATUS=OFF go run main.go -filter=full -v
> 
> # Clear environment variable
> unset FORCE_STATUS
> go run main.go  # Back to random
> ```
> 
> ## 4. **Validation cho Status Flag**
> 
> ````go
> func validateFlags() error {
> 	// Validate status if provided
> 	if *status != "" {
> 		validStatuses := []string{"ON", "OFF"}
> 		valid := false
> 		for _, validStatus := range validStatuses {
> 			if strings.ToUpper(*status) == validStatus {
> 				*status = strings.ToUpper(*status) // Normalize to uppercase
> 				valid = true
> 				break
> 			}
> 		}
> 		if !valid {
> 			return fmt.Errorf("invalid status '%s'. Valid options: ON, OFF", *status)
> 		}
> 	}
> 
> 	// ...existing validation code...
> 	return nil
> }
> ````
> 
> ## 5. **Usage Examples**
> 
> ### **Basic Usage:**
> ```bash
> # Test only ON servers
> go run main.go -status=ON
> 
> # Test only OFF servers  
> go run main.go -status=OFF
> 
> # Test with verbose output
> go run main.go -status=ON -v
> 
> # Combine with filter modes
> go run main.go -status=OFF -filter=env-only
> go run main.go -status=ON -filter=type-only -v
> ```
> 
> ### **Case Insensitive:**
> ```bash
> # These all work the same
> go run main.go -status=on
> go run main.go -status=ON
> go run main.go -status=On
> 
> go run main.go -status=off
> go run main.go -status=OFF
> go run main.go -status=Off
> ```
> 
> ### **With Environment Variables:**
> ```bash
> # Set once, use multiple times
> export FORCE_STATUS=ON
> go run main.go -filter=full
> go run main.go -filter=env-only
> go run main.go -filter=blank
> 
> # Override environment with flag
> FORCE_STATUS=OFF go run main.go -status=ON  # Will use ON (flag wins)
> ```
> 
> ## 6. **Updated Help Text**
> 
> ````go
> func showUsage() {
> 	fmt.Fprintf(os.Stderr, `
> 🔍 Server View Testing Tool
> 
> USAGE:
>   %s [FLAGS]
> 
> FLAGS:
> `, os.Args[0])
> 	flag.PrintDefaults()
> 
> 	fmt.Fprintf(os.Stderr, `
> EXAMPLES:
>   %s                           # Default: random status, full filtering
>   %s -status=ON                # Test only ON servers
>   %s -status=OFF -filter=blank # Test only OFF servers, no name filters
>   %s -status=ON -filter=env-only -v  # ON servers, env filtering, verbose
>   %s -help                     # Show detailed help
> 
> ENVIRONMENT VARIABLES:
>   FILTER_MODE    Override filter mode (same values as -filter flag)
>   FORCE_STATUS   Override status filter (ON, OFF)
> 
> `, os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0])
> }
> ````
> 
> ## 7. **Sample Output**
> 
> ```bash
> $ go run main.go -status=ON -v
> 🎯 Using specified status: ON
> 
> === Random Server View Tool ===
> 
> 🎲 Random View Parameters:
> ─────────────────────────────
>   📛 Filter by name: Database
>   🟢 Filter by status: ON
>   ⬆️ Sort by: name (asc)
>   📄 Pagination: from 1 to 12
> 
> 📋 Calling view API: http://localhost:8080/api/v1/servers?name=Database&status=ON&sort=name&order=asc&from=1&to=12
> ```
> 
> Bây giờ bạn có thể control status filtering trực tiếp từ terminal một cách linh hoạt!

tại sao default lại là fully mode

> Copilot: Lý do default được đặt là `full` mode có thể được giải thích qua một số yếu tố:
> 
> ## 1. **Testing Coverage tối đa**
> 
> ```go
> filterMode = flag.String("filter", "full", // ← Default = "full"
>     "Filter mode strategy:\n"+
>     "  • full      - All filters (types, environments, locations, partials)\n"+
>     "  • blank     - No name filters (only status/pagination)\n"+
>     // ...
> )
> ```
> 
> **Full mode** cung cấp **comprehensive testing** - test tất cả các loại filter có thể:
> - Server types (Web, Database, API...)
> - Environments (Production, Staging...)  
> - Locations (US-East, EU-Central...)
> - Partial matches (Server, Load, Data...)
> 
> ## 2. **Real-world Usage Simulation**
> 
> Trong thực tế, users thường sử dụng API với **nhiều loại filter khác nhau**:
> - Search theo tên server
> - Filter theo environment
> - Filter theo location
> - Filter theo server type
> 
> **Full mode** mô phỏng closest với actual usage patterns.
> 
> ## 3. **Comprehensive Validation**
> 
> ```go
> case "full", "complete":
>     filterNames = buildFullFilterNames()
>     // Tạo ra nhiều test cases nhất
> ```
> 
> Full mode giúp verify:
> - **All filter types** hoạt động correctly
> - **Combination của filters** work together
> - **Edge cases** với different filter values
> - **Performance** với diverse query patterns
> 
> ## 4. **Development Perspective**
> 
> Khi dev team chạy script lần đầu:
> ```bash
> go run main.go  # Không cần specify flags
> ```
> 
> Họ muốn **ngay lập tức** thấy tool test **toàn bộ functionality**, không chỉ basic features.
> 
> ## 5. **Progressive Testing Strategy**
> 
> ```
> Default (full) → Comprehensive testing
>      ↓
> Specific modes → Targeted testing
> ```
> 
> - **Start broad**: Full testing để catch major issues
> - **Narrow down**: Specific modes để debug particular features
> 
> ## 6. **Có thể thay đổi Default**
> 
> Nếu bạn muốn default khác, có thể modify:
> 
> ````go
> // Option 1: Change default to blank
> filterMode = flag.String("filter", "blank",  // ← Change default
>     "Filter mode strategy:\n"+
>     // ...
> )
> 
> // Option 2: Environment-based default
> func getDefaultFilterMode() string {
>     if env := os.Getenv("DEFAULT_FILTER_MODE"); env != "" {
>         return env
>     }
>     return "full"  // fallback
> }
> 
> var (
>     filterMode = flag.String("filter", getDefaultFilterMode(),
>         "Filter mode strategy:\n"+
>         // ...
>     )
> )
> ````
> 
> ## 7. **Alternative Approaches**
> 
> ### **No default (require explicit choice):**
> ```go
> filterMode = flag.String("filter", "",  // Empty default
>     "REQUIRED: Filter mode strategy:\n"+
>     // ...
> )
> 
> // In main():
> if *filterMode == "" {
>     fmt.Fprintf(os.Stderr, "Error: -filter flag is required\n")
>     flag.Usage()
>     os.Exit(1)
> }
> ```
> 
> ### **Smart default based on context:**
> ```go
> func getSmartDefault() string {
>     // Check if we're in CI/CD
>     if os.Getenv("CI") != "" {
>         return "blank"  // Simple testing in CI
>     }
>     
>     // Check if it's development
>     if os.Getenv("ENV") == "development" {
>         return "full"   // Comprehensive testing in dev
>     }
>     
>     return "blank"  // Conservative default
> }
> ```
> 
> ## 8. **Current Trade-offs**
> 
> ### **Pros của Full default:**
> - ✅ Maximum test coverage out-of-the-box
> - ✅ Discovers issues across all filter types
> - ✅ Realistic usage simulation
> 
> ### **Cons của Full default:**
> - ❌ Có thể overwhelm new users
> - ❌ Slower execution (more complex queries)
> - ❌ Harder to debug specific issues
> 
> ## 9. **Recommendation**
> 
> Để flexible hơn, có thể thêm **context-aware defaults**:
> 
> ````go
> func getContextualDefault() string {
>     // Check environment first
>     if mode := os.Getenv("DEFAULT_FILTER_MODE"); mode != "" {
>         return mode
>     }
>     
>     // Check if verbose flag is set (indicating debugging)
>     for _, arg := range os.Args {
>         if arg == "-v" || arg == "--verbose" {
>             return "blank"  // Simple mode for debugging
>         }
>     }
>     
>     // Default to full for comprehensive testing
>     return "full"
> }
> 
> var (
>     filterMode = flag.String("filter", getContextualDefault(),
>         "Filter mode strategy (default depends on context):\n"+
>         // ...
>     )
> )
> ````
> 
> **Usage:**
> ```bash
> # Full testing (default)
> go run main.go
> 
> # Debug mode (auto-detects verbose)
> go run main.go -v  # Will use 'blank' mode
> 
> # Override via environment
> DEFAULT_FILTER_MODE=blank go run main.go
> ```
> 
> Tóm lại, `full` làm default vì nó provide **maximum testing value** cho majority of use cases, nhưng có thể customize based on specific needs.
