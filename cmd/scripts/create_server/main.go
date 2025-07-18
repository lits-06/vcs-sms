package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
)

type CreateServerRequest struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	IPv4   string `json:"ipv4"`
	Status string `json:"status"`
}

var (
	serverTypes  = []string{"Web", "Database", "Cache", "Load Balancer", "API", "File", "Mail", "DNS", "Backup", "Monitor"}
	environments = []string{"Production", "Staging", "Development", "Testing", "Demo"}
	locations    = []string{"US-East", "US-West", "EU-Central", "Asia-Pacific", "Canada", "Australia"}
	statuses     = []string{"ON", "OFF"}
)

func main() {
	baseURL := "http://localhost:8080/api/v1/servers"

	numServers := 10
	concurrency := 3 // Số request đồng thời

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency)

	for i := 1; i <= numServers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			createServer(baseURL, id)
		}(i)
	}

	wg.Wait()
	fmt.Println("Đã tạo xong tất cả server!")
}

func createServer(baseURL string, id int) {
	req := CreateServerRequest{
		ID:     generateRandomID(),
		Name:   generateRandomName(),
		IPv4:   generateRandomIPv4(),
		Status: statuses[rand.Intn(len(statuses))],
	}

	jsonData, _ := json.Marshal(req)

	resp, err := http.Post(baseURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating server %d: %v", id, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 201 {
		fmt.Printf("✓ Created server %s\n", req.ID)
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("✗ Failed to create server %s: %s\n", req.ID, string(body))
	}
}

// Tạo ID ngẫu nhiên
func generateRandomID() string {
	prefixes := []string{"srv", "web", "db", "api", "lb", "cache", "file", "mail"}
	prefix := prefixes[rand.Intn(len(prefixes))]
	suffix := rand.Intn(9999) + 1
	return fmt.Sprintf("%s-%04d", prefix, suffix)
}

// Tạo tên server ngẫu nhiên
func generateRandomName() string {
	serverType := serverTypes[rand.Intn(len(serverTypes))]
	environment := environments[rand.Intn(len(environments))]
	location := locations[rand.Intn(len(locations))]
	number := rand.Intn(99) + 1

	return fmt.Sprintf("%s Server %02d (%s - %s)", serverType, number, environment, location)
}

// Tạo IPv4 ngẫu nhiên (trong dải private networks)
func generateRandomIPv4() string {
	// Sử dụng các dải IP private phổ biến
	networks := []string{
		"192.168.%d.%d", // 192.168.0.0/16
		"10.%d.%d.%d",   // 10.0.0.0/8
		"172.16.%d.%d",  // 172.16.0.0/12
	}

	network := networks[rand.Intn(len(networks))]

	switch network {
	case "192.168.%d.%d":
		return fmt.Sprintf(network, rand.Intn(256), rand.Intn(254)+1)
	case "10.%d.%d.%d":
		return fmt.Sprintf(network, rand.Intn(256), rand.Intn(256), rand.Intn(254)+1)
	case "172.16.%d.%d":
		return fmt.Sprintf(network, rand.Intn(16)+16, rand.Intn(254)+1)
	default:
		return fmt.Sprintf("192.168.%d.%d", rand.Intn(256), rand.Intn(254)+1)
	}
}
