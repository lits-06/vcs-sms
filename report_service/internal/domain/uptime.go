package domain

import "time"

// UptimeStats represents uptime statistics
type UptimeStats struct {
	StartDate        time.Time `json:"start_date"`
	EndDate          time.Time `json:"end_date"`
	TotalServers     int       `json:"total_servers"`
	OnlineServers    int       `json:"online_servers"`
	OfflineServers   int       `json:"offline_servers"`
	UptimePercentage float64   `json:"uptime_percentage"`
	ServerDetails    []ServerUptimeDetail `json:"server_details,omitempty"`
}

// ServerUptimeDetail represents uptime details for a specific server
type ServerUptimeDetail struct {
	ServerID         string    `json:"server_id"`
	UptimePercentage float64   `json:"uptime_percentage"`
}
