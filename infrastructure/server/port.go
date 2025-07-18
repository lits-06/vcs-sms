package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/lits-06/vcs-sms/entity"
)

// PortServerProvider implements ServerProvider interface
// This provider manages servers by starting/stopping services on specific ports
type PortServerProvider struct {
	activeServers map[string]*ServerProcess
}

// ServerProcess represents a running server process
type ServerProcess struct {
	ServerID   string
	Host       string
	Port       int
	HTTPServer *http.Server // HTTP server instance for Golang implementation
	Status     entity.ServerStatus
}

// NewPortServerProvider creates a new instance of PortServerProvider
func NewPortServerProvider() *PortServerProvider {
	return &PortServerProvider{
		activeServers: make(map[string]*ServerProcess),
	}
}

// CreateServer creates a new server by automatically finding an available port
func (p *PortServerProvider) CreateServer(ctx context.Context, server *entity.Server) error {
	// Find an available port automatically
	availablePort, err := p.findAvailablePort()
	if err != nil {
		return fmt.Errorf("failed to find available port: %w", err)
	}

	// Store server info but don't start it yet
	p.activeServers[server.ID] = &ServerProcess{
		ServerID:   server.ID,
		Host:       "localhost", // Always use localhost
		Port:       availablePort,
		HTTPServer: nil,
		Status:     entity.StatusOffline, // Always start as offline
	}

	// Follow server.Status - if ON, start the server automatically
	if server.Status == entity.StatusOnline {
		err = p.StartServer(ctx, server.ID)
		if err != nil {
			// If failed to start, clean up and return error
			delete(p.activeServers, server.ID)
			return fmt.Errorf("failed to start server after creation: %w", err)
		}
	}

	return nil
}

// UpdateServer updates an existing server's information
func (p *PortServerProvider) UpdateServer(ctx context.Context, server *entity.Server) error {
	serverProcess, exists := p.activeServers[server.ID]
	if !exists {
		return fmt.Errorf("server %s not found", server.ID)
	}

	// Get current status
	currentStatus := serverProcess.Status

	// Follow server.Status - handle status changes
	if server.Status == entity.StatusOnline && currentStatus == entity.StatusOffline {
		// Need to start the server
		err := p.StartServer(ctx, server.ID)
		if err != nil {
			return fmt.Errorf("failed to start server during update: %w", err)
		}
	} else if server.Status == entity.StatusOffline && currentStatus == entity.StatusOnline {
		// Need to stop the server
		err := p.StopServer(ctx, server.ID)
		if err != nil {
			return fmt.Errorf("failed to stop server during update: %w", err)
		}
	}
	// If status is the same, no action needed

	return nil
}

// DeleteServer stops and removes a server
func (p *PortServerProvider) DeleteServer(ctx context.Context, serverID string) error {
	serverProcess, exists := p.activeServers[serverID]
	if !exists {
		return fmt.Errorf("server %s not found", serverID)
	}

	// Stop the server if it's running
	if serverProcess.Status == entity.StatusOnline && serverProcess.HTTPServer != nil {
		err := p.stopServerProcess(serverProcess)
		if err != nil {
			return fmt.Errorf("Failed to stop server %s: %v", serverID, err)
		}
	}

	// Remove from active servers
	delete(p.activeServers, serverID)

	return nil
}

// StartServer starts a server on its designated port
func (p *PortServerProvider) StartServer(ctx context.Context, serverID string) error {
	serverProcess, exists := p.activeServers[serverID]
	if !exists {
		return fmt.Errorf("server %s not found", serverID)
	}

	if serverProcess.Status == entity.StatusOnline {
		return nil // Server is already running
	}

	// Check if port is available (it should be since we assigned it automatically)
	if p.isPortInUse(serverProcess.Port) {
		// If port is now in use, try to find a new one
		newPort, err := p.findAvailablePort()
		if err != nil {
			return fmt.Errorf("port %d is in use and no alternative port found: %w", serverProcess.Port, err)
		}
		serverProcess.Port = newPort
	}

	// Mark status as online before starting
	serverProcess.Status = entity.StatusOnline

	// Start a simple HTTP server on the port using Golang (simulation)
	// Create a simple HTTP server that serves on the specified port
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Server %s running on localhost:%d",
				serverID, serverProcess.Port)
		})
		mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{"server_id": "%s", "port": %d, "status": "%s"}`,
				serverID, serverProcess.Port, serverProcess.Status)
		})

		server := &http.Server{
			Addr:    fmt.Sprintf("localhost:%d", serverProcess.Port),
			Handler: mux,
		}

		// Store server instance for graceful shutdown
		serverProcess.HTTPServer = server

		// Start server (this will block)
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			serverProcess.Status = entity.StatusOffline
			serverProcess.HTTPServer = nil
		}
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Verify server is actually running
	if !p.isPortInUse(serverProcess.Port) {
		serverProcess.Status = entity.StatusOffline
		return fmt.Errorf("failed to start server on localhost:%d", serverProcess.Port)
	}

	return nil
}

// StopServer stops a running server
func (p *PortServerProvider) StopServer(ctx context.Context, serverID string) error {
	serverProcess, exists := p.activeServers[serverID]
	if !exists {
		return fmt.Errorf("server %s not found", serverID)
	}

	if serverProcess.Status == entity.StatusOffline {
		return nil // Server is already stopped
	}

	err := p.stopServerProcess(serverProcess)
	if err != nil {
		return fmt.Errorf("failed to stop server %s: %w", serverID, err)
	}

	return nil
}

// GetServerStatus returns the current status of a server by making HTTP health check
func (p *PortServerProvider) GetServerStatus(ctx context.Context, serverID string) (entity.ServerStatus, error) {
	serverProcess, exists := p.activeServers[serverID]
	if !exists {
		return entity.StatusOffline, fmt.Errorf("server %s not found", serverID)
	}

	// Make HTTP request to server's /status endpoint to check if it's really alive
	status := p.checkServerHealth(serverProcess)

	// Update the stored status
	serverProcess.Status = status

	// If server is not responding, clean up HTTPServer reference
	if status == entity.StatusOffline && serverProcess.HTTPServer != nil {
		serverProcess.HTTPServer = nil
	}

	return status, nil
}

// Helper methods
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

	// Double-check that the port is not already assigned to another server
	if p.isPortAssignedToServer(port) {
		// If already assigned, try to find another one
		return p.findPortStartingFrom(8001)
	}

	return port, nil
}

// checkServerHealth makes HTTP request to server to check if it's healthy
func (p *PortServerProvider) checkServerHealth(serverProcess *ServerProcess) entity.ServerStatus {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 3 * time.Second, // 3 second timeout
	}

	// Make request to server's status endpoint
	url := fmt.Sprintf("http://localhost:%d/status", serverProcess.Port)

	resp, err := client.Get(url)
	if err != nil {
		// Fallback: check if port is still in use
		if p.isPortInUse(serverProcess.Port) {
			// Port is in use but server not responding properly
			return entity.StatusOffline
		}
		return entity.StatusOffline
	}
	defer resp.Body.Close()

	// Check if response is successful
	if resp.StatusCode == http.StatusOK {
		return entity.StatusOnline
	}

	return entity.StatusOffline
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

		if !p.isPortInUse(port) && !p.isPortAssignedToServer(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available ports found after %d attempts starting from port %d", maxAttempts, startPort)
}

// isPortAssignedToServer checks if a port is already assigned to any managed server
func (p *PortServerProvider) isPortAssignedToServer(port int) bool {
	for _, serverProcess := range p.activeServers {
		if serverProcess.Port == port {
			return true
		}
	}
	return false
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

// stopServerProcess stops a server process
func (p *PortServerProvider) stopServerProcess(serverProcess *ServerProcess) error {
	// Stop HTTP server if it exists
	if serverProcess.HTTPServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := serverProcess.HTTPServer.Shutdown(ctx)
		if err != nil {
			// Force close if graceful shutdown fails
			serverProcess.HTTPServer.Close()
		}
		serverProcess.HTTPServer = nil
	}

	serverProcess.Status = entity.StatusOffline
	return nil
}
