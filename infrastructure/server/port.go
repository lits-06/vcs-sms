package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/lits-06/vcs-sms/entity"
	"github.com/redis/go-redis/v9"
)

const (
	ServerProcessKey = "server_process:%s"
)

// PortServerProvider implements ServerProvider interface
// This provider manages servers by starting/stopping services on specific ports
type PortServerProvider struct {
	redisClient *redis.Client
	httpServer  map[string]*http.Server // In-memory map of running servers
	mu          sync.RWMutex            // Mutex to protect access to httpServer
	// statusTicker *time.Ticker
	// stopChan     chan bool
}

// ServerProcess represents a running server process
type ServerProcess struct {
	ServerID string              `json:"server_id"`
	Host     string              `json:"host"`
	Port     int                 `json:"port"`
	Status   entity.ServerStatus `json:"status"`
}

// NewPortServerProvider creates a new PortServerProvider
func NewPortServerProvider(redisClient *redis.Client) *PortServerProvider {
	p := &PortServerProvider{
		redisClient: redisClient,
		httpServer:  make(map[string]*http.Server),
		// stopChan:    make(chan bool),
	}

	// Start periodic status reporting
	// p.startStatusReporting()

	return p
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
	p.forceStopServer(serverID)

	key := fmt.Sprintf(ServerProcessKey, serverID)
	if err := p.redisClient.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete server from Redis: %w", err)
	}

	return nil
}

// StartServer starts a server on its designated port
func (p *PortServerProvider) StartServer(ctx context.Context, serverID string) error {
	process, err := p.getServerProcess(ctx, serverID)
	if err != nil {
		return fmt.Errorf("server %s not found: %w", serverID, err)
	}

	p.mu.RLock()
	httpServer, exists := p.httpServer[serverID]
	p.mu.RUnlock()

	if exists && httpServer != nil {
		if p.checkServerHealth(process) == entity.StatusOnline {
			return nil // Server is already running and healthy
		}
		// Server object exists but not healthy, need to restart
		p.forceStopServer(serverID)
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

	p.mu.Lock()
	p.httpServer[serverID] = server
	p.mu.Unlock()
	// process.Status = entity.StatusOnline

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Failed to start server %s: %v\n", serverID, err)

			p.mu.Lock()
			delete(p.httpServer, serverID) // Remove from map if failed
			p.mu.Unlock()

			process.Status = entity.StatusOffline
			p.saveServerProcess(ctx, process)
		}
	}()

	// Wait a moment to ensure server started
	time.Sleep(200 * time.Millisecond)

	// Verify server is actually running
	if p.checkServerHealth(process) != entity.StatusOnline {
		p.forceStopServer(serverID)
		process.Status = entity.StatusOffline
		if err := p.saveServerProcess(ctx, process); err != nil {
			return fmt.Errorf("failed to save server process after start failure: %w", err)
		}
		return fmt.Errorf("server %s failed to start properly", serverID)
	}

	// Save updated process
	process.Status = entity.StatusOnline
	if err := p.saveServerProcess(ctx, process); err != nil {
		return fmt.Errorf("failed to save server process: %w", err)
	}

	return nil
}

// StopServer stops a running server
func (p *PortServerProvider) StopServer(ctx context.Context, serverID string) error {
	process, err := p.getServerProcess(ctx, serverID)
	if err != nil {
		return fmt.Errorf("server %s not found: %w", serverID, err)
	}

	// Force stop the server
	p.forceStopServer(serverID)

	if process.Status == entity.StatusOffline {
		return nil // Already stopped
	}

	process.Status = entity.StatusOffline

	// Save updated process
	if err := p.saveServerProcess(ctx, process); err != nil {
		return fmt.Errorf("failed to save server process: %w", err)
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

		if actualStatus == entity.StatusOffline {
			p.mu.Lock()
			delete(p.httpServer, serverID)
			p.mu.Unlock()
		}
	}

	return actualStatus, nil
}

// checkServerHealth makes HTTP request to server to check if it's healthy
func (p *PortServerProvider) checkServerHealth(serverProcess *ServerProcess) entity.ServerStatus {
	p.mu.RLock()
	httpServer, exists := p.httpServer[serverProcess.ServerID]
	p.mu.RUnlock()

	if !exists || httpServer == nil {
		return entity.StatusOffline // Server not running
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

	key := fmt.Sprintf(ServerProcessKey, process.ServerID)
	return p.redisClient.Set(ctx, key, data, 0).Err()
}

// getServerProcess retrieves server process from Redis
func (p *PortServerProvider) getServerProcess(ctx context.Context, serverID string) (*ServerProcess, error) {
	key := fmt.Sprintf(ServerProcessKey, serverID)
	data, err := p.redisClient.Get(ctx, key).Result()
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
	key := fmt.Sprintf(ServerProcessKey, serverID)
	exists, err := p.redisClient.Exists(ctx, key).Result()
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

	return port, nil
}

// Close stops the status reporting and cleans up resources
// func (p *PortServerProvider) Close() {
// 	if p.statusTicker != nil {
// 		p.statusTicker.Stop()
// 	}
// 	close(p.stopChan)
// }

// forceStopServer forcefully stops a server without updating Redis
func (p *PortServerProvider) forceStopServer(serverID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	httpServer, exists := p.httpServer[serverID]
	if !exists || httpServer == nil {
		return
	}

	// Try graceful shutdown first
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		// If graceful shutdown fails, try to close forcefully
		fmt.Printf("Graceful shutdown failed for server %s, forcing close: %v\n", serverID, err)
		httpServer.Close()
	}

	delete(p.httpServer, serverID)
}
