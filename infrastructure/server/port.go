package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/lits-06/vcs-sms/entity"
	"github.com/lits-06/vcs-sms/infrastructure/kafka"
	"github.com/redis/go-redis/v9"
)

// PortServerProvider implements ServerProvider interface
// This provider manages servers by starting/stopping services on specific ports
type PortServerProvider struct {
	redisClient  *redis.Client
	kafkaClient  *kafka.Client
	statusTicker *time.Ticker
	stopChan     chan bool
}

// ServerProcess represents a running server process
type ServerProcess struct {
	ServerID   string              `json:"server_id"`
	Host       string              `json:"host"`
	Port       int                 `json:"port"`
	HTTPServer *http.Server        `json:"-"`
	Status     entity.ServerStatus `json:"status"`
}

// NewPortServerProvider creates a new PortServerProvider
func NewPortServerProvider(kafkaClient *kafka.Client, redisClient *redis.Client) *PortServerProvider {
	p := &PortServerProvider{
		kafkaClient: kafkaClient,
		redisClient: redisClient,
		stopChan:    make(chan bool),
	}

	// Start periodic status reporting
	p.startStatusReporting()

	return p
}

// startStatusReporting starts periodic status reporting every 10 seconds
func (p *PortServerProvider) startStatusReporting() {
	p.statusTicker = time.NewTicker(10 * time.Second)

	go func() {
		for {
			select {
			case <-p.statusTicker.C:
				p.reportAllServerStatus()
			case <-p.stopChan:
				return
			}
		}
	}()
}

// reportAllServerStatus reports status of all managed servers
func (p *PortServerProvider) reportAllServerStatus() {
	ctx := context.Background()

	// Get all server keys from Redis
	keys, err := p.redisClient.Keys(ctx, "server:*").Result()
	if err != nil {
		fmt.Printf("Failed to get server keys from Redis: %v\n", err)
		return
	}

	for _, key := range keys {
		serverData, err := p.redisClient.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var process ServerProcess
		if err := json.Unmarshal([]byte(serverData), &process); err != nil {
			continue
		}

		// Check current health status
		currentStatus := p.checkServerHealth(&process)

		// Update status if changed
		if currentStatus != process.Status {
			process.Status = currentStatus
			p.saveServerProcess(ctx, &process)
		}

		// Publish status to Kafka
		statusMsg := kafka.ServerStatusMessage{
			ServerID:  process.ServerID,
			Status:    string(process.Status),
			Port:      process.Port,
			Host:      process.Host,
			Timestamp: time.Now(),
		}

		if err := p.kafkaClient.PublishServerStatus(ctx, statusMsg); err != nil {
			fmt.Printf("Failed to publish status for server %s: %v\n", process.ServerID, err)
		}
	}
}

// CreateServer creates a new server by automatically finding an available port
func (p *PortServerProvider) CreateServer(ctx context.Context, server *entity.Server) error {
	// Check if server already exists
	if p.serverExists(ctx, server.ID) {
		return fmt.Errorf("server %s already exists", server.ID)
	}

	// Find an available port automatically
	availablePort, err := p.findAvailablePort()
	if err != nil {
		return fmt.Errorf("failed to find available port: %w", err)
	}

	// Store server info but don't start it yet
	process := &ServerProcess{
		ServerID: server.ID,
		Host:     "localhost", // Always use localhost
		Port:     availablePort,
		Status:   entity.StatusOffline,
	}

	// Save to Redis
	if err := p.saveServerProcess(ctx, process); err != nil {
		return fmt.Errorf("failed to save server process: %w", err)
	}

	if server.Status == entity.StatusOnline {
		// If server is online, start it immediately
		if err := p.StartServer(ctx, server.ID); err != nil {
			return fmt.Errorf("failed to start server %s: %w", server.ID, err)
		}
	} else {
		// If server is offline, just save the initial status
		process.Status = server.Status
		if err := p.saveServerProcess(ctx, process); err != nil {
			return fmt.Errorf("failed to save initial server status: %w", err)
		}
	}

	finalProcess, _ := p.getServerProcess(ctx, server.ID)

	// Publish initial status immediately
	statusMsg := kafka.ServerStatusMessage{
		ServerID:  server.ID,
		Status:    string(finalProcess.Status),
		Port:      finalProcess.Port,
		Host:      finalProcess.Host,
		Timestamp: time.Now(),
	}

	if err := p.kafkaClient.PublishServerStatus(ctx, statusMsg); err != nil {
		fmt.Printf("Failed to publish initial status for server %s: %v\n", server.ID, err)
	}

	return nil
}

// UpdateServer updates an existing server's information
func (p *PortServerProvider) UpdateServer(ctx context.Context, server *entity.Server) error {
	_, err := p.getServerProcess(ctx, server.ID)
	if err != nil {
		return fmt.Errorf("server %s not found: %w", server.ID, err)
	}

	if server.Status == entity.StatusOnline {
		// If server is online, start it
		if err := p.StartServer(ctx, server.ID); err != nil {
			return fmt.Errorf("failed to start server %s: %w", server.ID, err)
		}
	} else {
		// If server is offline, stop it
		if err := p.StopServer(ctx, server.ID); err != nil {
			return fmt.Errorf("failed to stop server %s: %w", server.ID, err)
		}
	}

	return nil
}

// DeleteServer stops and removes a server
func (p *PortServerProvider) DeleteServer(ctx context.Context, serverID string) error {
	process, err := p.getServerProcess(ctx, serverID)
	if err != nil {
		return fmt.Errorf("server %s not found: %w", serverID, err)
	}

	// Stop the server if it's running
	if process.HTTPServer != nil {
		if err := process.HTTPServer.Shutdown(ctx); err != nil {
			fmt.Printf("Failed to shutdown server %s gracefully: %v\n", serverID, err)
		}
	}

	if err := p.redisClient.Del(ctx, "server:"+serverID).Err(); err != nil {
		return fmt.Errorf("failed to delete server from Redis: %w", err)
	}

	// Publish deletion status
	statusMsg := kafka.ServerStatusMessage{
		ServerID:  serverID,
		Status:    string(entity.StatusOffline),
		Port:      process.Port,
		Host:      process.Host,
		Timestamp: time.Now(),
	}

	if err := p.kafkaClient.PublishServerStatus(ctx, statusMsg); err != nil {
		fmt.Printf("Failed to publish deletion status for server %s: %v\n", serverID, err)
	}

	return nil
}

// StartServer starts a server on its designated port
func (p *PortServerProvider) StartServer(ctx context.Context, serverID string) error {
	process, err := p.getServerProcess(ctx, serverID)
	if err != nil {
		return fmt.Errorf("server %s not found: %w", serverID, err)
	}

	if process.HTTPServer != nil {
		actualStatus := p.checkServerHealth(process)
		if actualStatus == entity.StatusOnline {
			return nil // Server is already running and healthy
		}
		// Server object exists but not healthy, need to restart
		p.forceStopServer(process)
	}

	// Create a simple HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","server_id":"` + serverID + `", "port":` + fmt.Sprintf("%d", process.Port) + `, "host":"` + process.Host + `"}`))
	})

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", process.Host, process.Port),
		Handler: mux,
	}

	process.HTTPServer = server
	process.Status = entity.StatusOnline

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Failed to start server %s: %v\n", serverID, err)
			process.Status = entity.StatusOffline
			process.HTTPServer = nil
			p.saveServerProcess(ctx, process)
		}
	}()

	// Wait a moment to ensure server started
	time.Sleep(200 * time.Millisecond)

	// Verify server is actually running
	actualStatus := p.checkServerHealth(process)
	if actualStatus != entity.StatusOnline {
		process.Status = entity.StatusOffline
		process.HTTPServer = nil
		if err := p.saveServerProcess(ctx, process); err != nil {
			return fmt.Errorf("failed to save server process after start failure: %w", err)
		}
		return fmt.Errorf("server %s failed to start properly", serverID)
	}

	// Save updated process
	if err := p.saveServerProcess(ctx, process); err != nil {
		return fmt.Errorf("failed to save server process: %w", err)
	}

	// Publish start status
	statusMsg := kafka.ServerStatusMessage{
		ServerID:  serverID,
		Status:    string(process.Status),
		Port:      process.Port,
		Host:      process.Host,
		Timestamp: time.Now(),
	}

	if err := p.kafkaClient.PublishServerStatus(ctx, statusMsg); err != nil {
		fmt.Printf("Failed to publish start status for server %s: %v\n", serverID, err)
	}

	return nil
}

// StopServer stops a running server
func (p *PortServerProvider) StopServer(ctx context.Context, serverID string) error {
	process, err := p.getServerProcess(ctx, serverID)
	if err != nil {
		return fmt.Errorf("server %s not found: %w", serverID, err)
	}

	if process.HTTPServer == nil && process.Status == entity.StatusOffline {
		return nil
	}

	// Force stop the server
	p.forceStopServer(process)

	process.HTTPServer = nil
	process.Status = entity.StatusOffline

	// Save updated process
	if err := p.saveServerProcess(ctx, process); err != nil {
		return fmt.Errorf("failed to save server process: %w", err)
	}

	// Publish stop status
	statusMsg := kafka.ServerStatusMessage{
		ServerID:  serverID,
		Status:    string(process.Status),
		Port:      process.Port,
		Host:      process.Host,
		Timestamp: time.Now(),
	}

	if err := p.kafkaClient.PublishServerStatus(ctx, statusMsg); err != nil {
		fmt.Printf("Failed to publish stop status for server %s: %v\n", serverID, err)
	}

	return nil
}

// GetServerStatus returns the current status of a server by making HTTP health check
func (p *PortServerProvider) GetServerStatus(ctx context.Context, serverID string) (entity.ServerStatus, error) {
	process, err := p.getServerProcess(ctx, serverID)
	if err != nil {
		return entity.StatusOffline, fmt.Errorf("server %s not found: %w", serverID, err)
	}

	// Check actual health status
	actualStatus := p.checkServerHealth(process)

	// Update status if different
	if actualStatus != process.Status {
		process.Status = actualStatus
		p.saveServerProcess(ctx, process)
	}

	return actualStatus, nil
}

// checkServerHealth makes HTTP request to server to check if it's healthy
func (p *PortServerProvider) checkServerHealth(serverProcess *ServerProcess) entity.ServerStatus {
	if serverProcess.HTTPServer == nil {
		return entity.StatusOffline
	}

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://%s:%d/health", serverProcess.Host, serverProcess.Port))
	if err != nil {
		return entity.StatusOffline
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return entity.StatusOnline
	}

	return entity.StatusOffline
}

// saveServerProcess saves server process to Redis
func (p *PortServerProvider) saveServerProcess(ctx context.Context, process *ServerProcess) error {
	data, err := json.Marshal(process)
	if err != nil {
		return err
	}

	return p.redisClient.Set(ctx, "server:"+process.ServerID, data, 0).Err()
}

// getServerProcess retrieves server process from Redis
func (p *PortServerProvider) getServerProcess(ctx context.Context, serverID string) (*ServerProcess, error) {
	data, err := p.redisClient.Get(ctx, "server:"+serverID).Result()
	if err != nil {
		return nil, err
	}

	var process ServerProcess
	if err := json.Unmarshal([]byte(data), &process); err != nil {
		return nil, err
	}

	return &process, nil
}

// serverExists checks if server exists in Redis
func (p *PortServerProvider) serverExists(ctx context.Context, serverID string) bool {
	exists, err := p.redisClient.Exists(ctx, "server:"+serverID).Result()
	return err == nil && exists > 0
}

func (p *PortServerProvider) findAvailablePort() (int, error) {
	// Try to let OS assign an available port by listening on port 0
	// This is much faster than looping through ports
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, fmt.Errorf("failed to find available port: %w", err)
	}
	defer listener.Close()

	// Get the assigned port
	addr := listener.Addr().(*net.TCPAddr)
	port := addr.Port

	// Ensure port is after 8000 (main program port)
	if port <= 8000 {
		// If OS assigned port <= 8000, try to find a port manually starting from 8001
		return p.findPortStartingFrom(8001)
	}

	return port, nil
}

// findPortStartingFrom finds available port starting from specified port (fallback method)
func (p *PortServerProvider) findPortStartingFrom(startPort int) (int, error) {
	// Only check a reasonable range to avoid infinite loops
	maxAttempts := 100
	for i := 0; i < maxAttempts; i++ {
		port := startPort + i
		if port > 65535 {
			break
		}

		if !p.isPortInUse(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available ports found after %d attempts starting from port %d", maxAttempts, startPort)
}

// isPortInUse checks if a port is currently in use on localhost
func (p *PortServerProvider) isPortInUse(port int) bool {
	conn, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return true // Port is in use
	}
	conn.Close()
	return false // Port is available
}

// Close stops the status reporting and cleans up resources
func (p *PortServerProvider) Close() {
	if p.statusTicker != nil {
		p.statusTicker.Stop()
	}
	close(p.stopChan)
}

// forceStopServer forcefully stops a server without updating Redis
func (p *PortServerProvider) forceStopServer(process *ServerProcess) {
	if process.HTTPServer == nil {
		return
	}

	// Try graceful shutdown first
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := process.HTTPServer.Shutdown(shutdownCtx); err != nil {
		// If graceful shutdown fails, try to close forcefully
		fmt.Printf("Graceful shutdown failed for server %s, forcing close: %v\n", process.ServerID, err)
		process.HTTPServer.Close()
	}
}
