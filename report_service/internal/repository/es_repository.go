package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/lits-06/vcs-sms/report_service/config"
	"github.com/lits-06/vcs-sms/report_service/internal/domain"
)

type recordRepository struct {
	esClient *elasticsearch.Client
	snapshotIdx string
	recordIdx   string
}

func NewRecordRepository(esClient *elasticsearch.Client, cfg *config.Config) domain.Repository {
	return &recordRepository{
		esClient: esClient,
		snapshotIdx: cfg.Elasticsearch.SnapshotIndex,
		recordIdx:   cfg.Elasticsearch.RecordIndex,
	}
}

// GetUptimeStats calculates comprehensive uptime statistics for all servers
func (r *recordRepository) GetUptimeStats(ctx context.Context, startDate, endDate time.Time) (*domain.UptimeStats, error) {
	// Calculate total time period in hours
	TotalSeconds := endDate.Sub(startDate).Seconds()

	// Get all server snapshots to know total servers
	serverSnapshots, err := r.getAllServerSnapshots(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get server snapshots: %w", err)
	}
	
	totalServers := len(serverSnapshots)
	if totalServers == 0 {
		return &domain.UptimeStats{
			StartDate:          startDate,
			EndDate:            endDate,
			TotalServers:       0,
			OnlineServers:      0,
			OfflineServers:     0,
			UptimePercentage:   0,
		}, nil
	}

	uptimeDetail := make([]domain.ServerUptimeDetail, 0, totalServers)

	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, server := range serverSnapshots {
		wg.Add(1)
		go func(s domain.Server) {
			defer wg.Done()
			event, err := r.getServerEvents(ctx, s.ServerID, startDate, endDate)
			if err != nil {
				// Log error and continue
				fmt.Printf("Error fetching events for server %s: %v\n", s.ServerID, err)
				return
			}

			previousStatus, err := r.lastStatusBeforeDate(ctx, s.ServerID, startDate)
			if err != nil {
				// Log error and continue
				fmt.Printf("Error fetching last status for server %s: %v\n", s.ServerID, err)
				return
			}

			uptimeSeconds := r.calculateUptimeSeconds(event, startDate, endDate, previousStatus)
			detail := domain.ServerUptimeDetail{
				ServerID:         s.ServerID,
				UptimePercentage: float64(uptimeSeconds) / TotalSeconds * 100,
			}

			mu.Lock()
			uptimeDetail = append(uptimeDetail, detail)
			mu.Unlock()
		}(server)
	}
	wg.Wait()
	
	var onlineServers int
	var totalUptimePercentage float64
	for _, detail := range uptimeDetail {
		if serverSnapshots[detail.ServerID].Status == "ON" {
			onlineServers++
		}
		totalUptimePercentage += detail.UptimePercentage
	}
	
	return &domain.UptimeStats{
		StartDate:          startDate,
		EndDate:            endDate,
		TotalServers:       totalServers,
		OnlineServers:      onlineServers,
		OfflineServers:     totalServers - onlineServers,
		UptimePercentage:   totalUptimePercentage / float64(totalServers),
		ServerDetails:      uptimeDetail,
	}, nil
}

// Helper functions
func (r *recordRepository) getAllServerSnapshots(ctx context.Context) (map[string]domain.Server, error) {
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
		r.esClient.Search.WithIndex(r.snapshotIdx),
		r.esClient.Search.WithBody(bytes.NewReader(queryBytes)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search snapshots: %w", err)
	}
	defer res.Body.Close()
	
	var searchResponse struct {
		Hits struct {
			Hits []struct {
				Source domain.Server `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	
	if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	result := make(map[string]domain.Server)
	for _, hit := range searchResponse.Hits.Hits {
		result[hit.Source.ServerID] = hit.Source
	}
	
	return result, nil
}

func (r *recordRepository) getServerEvents(ctx context.Context, serverID string, startDate, endDate time.Time) ([]domain.Server, error) {
	query := map[string]interface{}{
		"size": 10000,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"server_id": serverID,
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
		r.esClient.Search.WithIndex(r.recordIdx),
		r.esClient.Search.WithBody(bytes.NewReader(queryBytes)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search events: %w", err)
	}
	defer res.Body.Close()
	
	var searchResponse struct {
		Hits struct {
			Hits []struct {
				Source domain.Server `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	
	if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	result := make([]domain.Server, 0, len(searchResponse.Hits.Hits))
	for _, hit := range searchResponse.Hits.Hits {
		result = append(result, hit.Source)
	}
	
	return result, nil
}

func (r *recordRepository) calculateUptimeSeconds(events []domain.Server, startDate, endDate time.Time, previousStatus string) int64 {
	var uptimeSeconds int64
	var lastOnlineTime *time.Time
	
	if len(events) == 0 {
		if previousStatus == "ON" {
			return int64(endDate.Sub(startDate).Seconds())
		}
		return 0
	}

	if len(events) == 1 && events[0].Status == "OFF" {
		if previousStatus == "ON" {
			return int64(events[0].Timestamp.Sub(startDate).Seconds())
		}

		return 0
	}	
	
	for _, event := range events {
		if event.Status == "ON" {
			lastOnlineTime = &event.Timestamp
		} else if event.Status == "OFF" && lastOnlineTime != nil {
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
		if onlineStart.Before(endDate) {
			uptimeSeconds += int64(endDate.Sub(onlineStart).Seconds())
		}
	}
	
	return uptimeSeconds
}

func (r *recordRepository) lastStatusBeforeDate(ctx context.Context, serverID string, startDate time.Time) (string, error) {
    query := map[string]interface{}{
        "size": 1,
        "query": map[string]interface{}{
            "bool": map[string]interface{}{
                "must": []map[string]interface{}{
                    {
                        "term": map[string]interface{}{
                            "server_id": serverID,
                        },
                    },
                    {
                        "range": map[string]interface{}{
                            "timestamp": map[string]interface{}{
                                "lt": startDate.Format(time.RFC3339),
                            },
                        },
                    },
                },
            },
        },
        "sort": []map[string]interface{}{
            {
                "timestamp": map[string]string{"order": "desc"},
            },
        },
    }
    
    queryBytes, err := json.Marshal(query)
    if err != nil {
        return "OFF", fmt.Errorf("failed to marshal query: %w", err)
    }
    
    res, err := r.esClient.Search(
        r.esClient.Search.WithContext(ctx),
        r.esClient.Search.WithIndex(r.recordIdx),
        r.esClient.Search.WithBody(bytes.NewReader(queryBytes)),
    )
    if err != nil {
        return "OFF", fmt.Errorf("failed to search events: %w", err)
    }
    defer res.Body.Close()
    
    var searchResponse struct {
        Hits struct {
            Total struct {
                Value int64 `json:"value"`
            } `json:"total"`
            Hits []struct {
                Source domain.Server `json:"_source"`
            } `json:"hits"`
        } `json:"hits"`
    }
    
    if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
        return "OFF", fmt.Errorf("failed to decode response: %w", err)
    }
    
    if searchResponse.Hits.Total.Value == 0 {
        return "OFF", nil
    }

    return searchResponse.Hits.Hits[0].Source.Status, nil
}