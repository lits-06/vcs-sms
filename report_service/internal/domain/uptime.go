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
	TotalUptimeHours float64   `json:"total_uptime_hours"`
	TotalPossibleHours float64 `json:"total_possible_hours"`
	ServerDetails    []ServerUptimeDetail `json:"server_details,omitempty"`
}

// ServerUptimeDetail represents uptime details for a specific server
type ServerUptimeDetail struct {
	ServerID         string    `json:"server_id"`
	UptimePercentage float64   `json:"uptime_percentage"`
	UptimeHours      float64   `json:"uptime_hours"`
	DowntimeHours    float64   `json:"downtime_hours"`
	LastSeen         time.Time `json:"last_seen"`
	CurrentStatus    string    `json:"current_status"`
}
