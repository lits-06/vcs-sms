package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/lits-06/vcs-sms/report_service/internal/domain"
)

type recordRepository struct {
	esClient *elasticsearch.Client
}

func NewRecordRepository(esClient *elasticsearch.Client) domain.Repository {
	return &recordRepository{esClient: esClient}
}

// GetUptimeStats calculates comprehensive uptime statistics for all servers
func (r *recordRepository) GetUptimeStats(ctx context.Context, req *domain.UptimeRequest) (*domain.UptimeStats, error) {
	// Calculate total time period in hours
	totalHours := req.EndDate.Sub(req.StartDate).Hours()
	
	// Get all server snapshots to know total servers
	serverSnapshots, err := r.getAllServerSnapshots(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get server snapshots: %w", err)
	}
	
	totalServers := len(serverSnapshots)
	if totalServers == 0 {
		return &domain.UptimeStats{
			StartDate:          req.StartDate,
			EndDate:            req.EndDate,
			TotalServers:       0,
			OnlineServers:      0,
			OfflineServers:     0,
			UptimePercentage:   0,
			TotalUptimeHours:   0,
			TotalPossibleHours: 0,
		}, nil
	}
	
	// Get uptime statistics for all servers
	serverIDs := make([]string, 0, totalServers)
	for serverID := range serverSnapshots {
		serverIDs = append(serverIDs, serverID)
	}
	
	serverDetails, err := r.GetServerUptimeStats(ctx, serverIDs, req.StartDate, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get server uptime stats: %w", err)
	}
	
	// Calculate aggregated statistics
	var totalUptimeHours float64
	var onlineServers int
	totalPossibleHours := float64(totalServers) * totalHours
	
	for _, detail := range serverDetails {
		totalUptimeHours += detail.UptimeHours
		if detail.CurrentStatus == "online" {
			onlineServers++
		}
	}
	
	uptimePercentage := 0.0
	if totalPossibleHours > 0 {
		uptimePercentage = (totalUptimeHours / totalPossibleHours) * 100
	}
	
	return &domain.UptimeStats{
		StartDate:          req.StartDate,
		EndDate:            req.EndDate,
		TotalServers:       totalServers,
		OnlineServers:      onlineServers,
		OfflineServers:     totalServers - onlineServers,
		UptimePercentage:   uptimePercentage,
		TotalUptimeHours:   totalUptimeHours,
		TotalPossibleHours: totalPossibleHours,
		ServerDetails:      serverDetails,
	}, nil
}

// GetServerUptimeStats calculates uptime for specific servers
func (r *recordRepository) GetServerUptimeStats(ctx context.Context, serverIDs []string, startDate, endDate time.Time) ([]domain.ServerUptimeDetail, error) {
	query := map[string]interface{}{
		"size": 10000,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"terms": map[string]interface{}{
							"server_id": serverIDs,
						},
					},
					{
						"range": map[string]interface{}{
							"timestamp": map[string]interface{}{
								"gte": startDate.Format(time.RFC3339),
								"lte": endDate.Format(time.RFC3339),
							},
						},
					},
				},
			},
		},
		"sort": []map[string]interface{}{
			{
				"server_id": map[string]string{"order": "asc"},
			},
			{
				"timestamp": map[string]string{"order": "asc"},
			},
		},
	}
	
	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}
	
	// Search state events
	res, err := r.esClient.Search(
		r.esClient.Search.WithContext(ctx),
		r.esClient.Search.WithIndex("server_state_events"),
		r.esClient.Search.WithBody(bytes.NewReader(queryBytes)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search events: %w", err)
	}
	defer res.Body.Close()
	
	var searchResponse struct {
		Hits struct {
			Hits []struct {
				Source domain.ServerStateRecord `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	
	if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Get current server status
	currentStatus, err := r.GetServerCurrentStatus(ctx, serverIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get current status: %w", err)
	}
	
	// Calculate uptime for each server
	return r.calculateServerUptime(searchResponse.Hits.Hits, currentStatus, startDate, endDate), nil
}

// GetServerCurrentStatus gets current status of servers
func (r *recordRepository) GetServerCurrentStatus(ctx context.Context, serverIDs []string) (map[string]domain.ServerSnapshot, error) {
	query := map[string]interface{}{
		"size": len(serverIDs),
		"query": map[string]interface{}{
			"terms": map[string]interface{}{
				"server_id": serverIDs,
			},
		},
	}
	
	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}
	
	res, err := r.esClient.Search(
		r.esClient.Search.WithContext(ctx),
		r.esClient.Search.WithIndex("server_snapshots"),
		r.esClient.Search.WithBody(bytes.NewReader(queryBytes)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search snapshots: %w", err)
	}
	defer res.Body.Close()
	
	var searchResponse struct {
		Hits struct {
			Hits []struct {
				Source domain.ServerSnapshot `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	
	if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	result := make(map[string]domain.ServerSnapshot)
	for _, hit := range searchResponse.Hits.Hits {
		result[hit.Source.ServerID] = hit.Source
	}
	
	return result, nil
}

// GetServerStateEvents gets state change events for servers
func (r *recordRepository) GetServerStateEvents(ctx context.Context, serverIDs []string, startDate, endDate time.Time) ([]domain.ServerStateRecord, error) {
	query := map[string]interface{}{
		"size": 10000,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"terms": map[string]interface{}{
							"server_id": serverIDs,
						},
					},
					{
						"range": map[string]interface{}{
							"timestamp": map[string]interface{}{
								"gte": startDate.Format(time.RFC3339),
								"lte": endDate.Format(time.RFC3339),
							},
						},
					},
				},
			},
		},
		"sort": []map[string]interface{}{
			{
				"timestamp": map[string]string{"order": "asc"},
			},
		},
	}
	
	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}
	
	res, err := r.esClient.Search(
		r.esClient.Search.WithContext(ctx),
		r.esClient.Search.WithIndex("server_state_events"),
		r.esClient.Search.WithBody(bytes.NewReader(queryBytes)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search events: %w", err)
	}
	defer res.Body.Close()
	
	var searchResponse struct {
		Hits struct {
			Hits []struct {
				Source domain.ServerStateRecord `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	
	if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	result := make([]domain.ServerStateRecord, 0, len(searchResponse.Hits.Hits))
	for _, hit := range searchResponse.Hits.Hits {
		result = append(result, hit.Source)
	}
	
	return result, nil
}

// Helper functions

func (r *recordRepository) getAllServerSnapshots(ctx context.Context) (map[string]domain.ServerSnapshot, error) {
	query := map[string]interface{}{
		"size": 10000,
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}
	
	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}
	
	res, err := r.esClient.Search(
		r.esClient.Search.WithContext(ctx),
		r.esClient.Search.WithIndex("server_snapshots"),
		r.esClient.Search.WithBody(bytes.NewReader(queryBytes)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search snapshots: %w", err)
	}
	defer res.Body.Close()
	
	var searchResponse struct {
		Hits struct {
			Hits []struct {
				Source domain.ServerSnapshot `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	
	if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	result := make(map[string]domain.ServerSnapshot)
	for _, hit := range searchResponse.Hits.Hits {
		result[hit.Source.ServerID] = hit.Source
	}
	
	return result, nil
}

func (r *recordRepository) calculateServerUptime(events []struct {
	Source domain.ServerStateRecord `json:"_source"`
}, currentStatus map[string]domain.ServerSnapshot, startDate, endDate time.Time) []domain.ServerUptimeDetail {
	
	serverUptime := make(map[string]*domain.ServerUptimeDetail)
	serverEvents := make(map[string][]domain.ServerStateRecord)
	
	// Group events by server
	for _, event := range events {
		serverID := event.Source.ServerID
		if _, exists := serverEvents[serverID]; !exists {
			serverEvents[serverID] = make([]domain.ServerStateRecord, 0)
		}
		serverEvents[serverID] = append(serverEvents[serverID], event.Source)
	}
	
	totalHours := endDate.Sub(startDate).Hours()
	
	// Calculate uptime for each server
	for serverID, events := range serverEvents {
		detail := &domain.ServerUptimeDetail{
			ServerID:      serverID,
			UptimeHours:   0,
			DowntimeHours: 0,
		}
		
		// Get current status
		if snapshot, exists := currentStatus[serverID]; exists {
			detail.CurrentStatus = snapshot.CurrentStatus
			detail.LastSeen = snapshot.LastUpdated
		}
		
		// Calculate uptime based on state transitions
		uptimeSeconds := r.calculateUptimeSeconds(events, startDate, endDate, detail.CurrentStatus)
		detail.UptimeHours = float64(uptimeSeconds) / 3600
		detail.DowntimeHours = totalHours - detail.UptimeHours
		
		if totalHours > 0 {
			detail.UptimePercentage = (detail.UptimeHours / totalHours) * 100
		}
		
		serverUptime[serverID] = detail
	}
	
	// Add servers that have no events in the time period
	for serverID, snapshot := range currentStatus {
		if _, exists := serverUptime[serverID]; !exists {
			detail := &domain.ServerUptimeDetail{
				ServerID:         serverID,
				CurrentStatus:    snapshot.CurrentStatus,
				LastSeen:         snapshot.LastUpdated,
				UptimePercentage: 0,
				UptimeHours:      0,
				DowntimeHours:    totalHours,
			}
			
			// If server was online before start date, assume it was online for the entire period
			if snapshot.CurrentStatus == "online" && snapshot.LastOnline.Before(startDate) {
				detail.UptimeHours = totalHours
				detail.DowntimeHours = 0
				detail.UptimePercentage = 100
			}
			
			serverUptime[serverID] = detail
		}
	}
	
	// Convert to slice
	result := make([]domain.ServerUptimeDetail, 0, len(serverUptime))
	for _, detail := range serverUptime {
		result = append(result, *detail)
	}
	
	return result
}

func (r *recordRepository) calculateUptimeSeconds(events []domain.ServerStateRecord, startDate, endDate time.Time, currentStatus string) int64 {
	var uptimeSeconds int64
	var lastOnlineTime *time.Time
	
	// Sort events by timestamp
	// Events should already be sorted from ES query
	
	for _, event := range events {
		if event.Status == "online" {
			lastOnlineTime = &event.Timestamp
		} else if event.Status == "offline" && lastOnlineTime != nil {
			// Calculate uptime for this online period
			onlineStart := *lastOnlineTime
			offlineTime := event.Timestamp
			
			// Adjust for query time range
			if onlineStart.Before(startDate) {
				onlineStart = startDate
			}
			if offlineTime.After(endDate) {
				offlineTime = endDate
			}
			
			if onlineStart.Before(offlineTime) {
				uptimeSeconds += int64(offlineTime.Sub(onlineStart).Seconds())
			}
			
			lastOnlineTime = nil
		}
	}
	
	// If server is still online at the end of the period
	if lastOnlineTime != nil {
		onlineStart := *lastOnlineTime
		if onlineStart.Before(startDate) {
			onlineStart = startDate
		}
		
		if onlineStart.Before(endDate) {
			uptimeSeconds += int64(endDate.Sub(onlineStart).Seconds())
		}
	}
	
	return uptimeSeconds
}