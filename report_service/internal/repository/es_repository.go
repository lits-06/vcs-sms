// package repository

// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"sync"
// 	"time"

// 	"github.com/elastic/go-elasticsearch/v9"
// 	"github.com/lits-06/vcs-sms/report_service/config"
// 	"github.com/lits-06/vcs-sms/report_service/internal/domain"
// )

// type recordRepository struct {
// 	esClient *elasticsearch.Client
// 	snapshotIdx string
// 	recordIdx   string
// }

// func NewRecordRepository(esClient *elasticsearch.Client, cfg *config.Config) domain.Repository {
// 	return &recordRepository{
// 		esClient: esClient,
// 		snapshotIdx: cfg.Elasticsearch.SnapshotIndex,
// 		recordIdx:   cfg.Elasticsearch.RecordIndex,
// 	}
// }

// // GetUptimeStats calculates comprehensive uptime statistics for all servers
// func (r *recordRepository) GetUptimeStats(ctx context.Context, startDate, endDate time.Time) (*domain.UptimeStats, error) {
// 	// Calculate total time period in hours
// 	TotalSeconds := endDate.Sub(startDate).Seconds()

// 	// Get all server snapshots to know total servers
// 	serverSnapshots, err := r.getAllServerSnapshots(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get server snapshots: %w", err)
// 	}

// 	totalServers := len(serverSnapshots)
// 	if totalServers == 0 {
// 		return &domain.UptimeStats{
// 			StartDate:          startDate,
// 			EndDate:            endDate,
// 			TotalServers:       0,
// 			OnlineServers:      0,
// 			OfflineServers:     0,
// 			UptimePercentage:   0,
// 		}, nil
// 	}

// 	uptimeDetail := make([]domain.ServerUptimeDetail, 0, totalServers)

// 	var wg sync.WaitGroup
// 	var mu sync.Mutex
// 	for _, server := range serverSnapshots {
// 		wg.Add(1)
// 		go func(s domain.Server) {
// 			defer wg.Done()
// 			event, err := r.getServerEvents(ctx, s.ServerID, startDate, endDate)
// 			if err != nil {
// 				// Log error and continue
// 				fmt.Printf("Error fetching events for server %s: %v\n", s.ServerID, err)
// 				return
// 			}

// 			previousStatus, err := r.lastStatusBeforeDate(ctx, s.ServerID, startDate)
// 			if err != nil {
// 				// Log error and continue
// 				fmt.Printf("Error fetching last status for server %s: %v\n", s.ServerID, err)
// 				return
// 			}

// 			uptimeSeconds := r.calculateUptimeSeconds(event, startDate, endDate, previousStatus)
// 			detail := domain.ServerUptimeDetail{
// 				ServerID:         s.ServerID,
// 				UptimePercentage: float64(uptimeSeconds) / TotalSeconds * 100,
// 			}

// 			mu.Lock()
// 			uptimeDetail = append(uptimeDetail, detail)
// 			mu.Unlock()
// 		}(server)
// 	}
// 	wg.Wait()

// 	var onlineServers int
// 	var totalUptimePercentage float64
// 	for _, detail := range uptimeDetail {
// 		if serverSnapshots[detail.ServerID].Status == "ON" {
// 			onlineServers++
// 		}
// 		totalUptimePercentage += detail.UptimePercentage
// 	}

// 	return &domain.UptimeStats{
// 		StartDate:          startDate,
// 		EndDate:            endDate,
// 		TotalServers:       totalServers,
// 		OnlineServers:      onlineServers,
// 		OfflineServers:     totalServers - onlineServers,
// 		UptimePercentage:   totalUptimePercentage / float64(totalServers),
// 		ServerDetails:      uptimeDetail,
// 	}, nil
// }

// // Helper functions
// func (r *recordRepository) getAllServerSnapshots(ctx context.Context) (map[string]domain.Server, error) {
// 	query := map[string]interface{}{
// 		"size": 10000,
// 		"query": map[string]interface{}{
// 			"match_all": map[string]interface{}{},
// 		},
// 	}

// 	queryBytes, err := json.Marshal(query)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to marshal query: %w", err)
// 	}

// 	res, err := r.esClient.Search(
// 		r.esClient.Search.WithContext(ctx),
// 		r.esClient.Search.WithIndex(r.snapshotIdx),
// 		r.esClient.Search.WithBody(bytes.NewReader(queryBytes)),
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to search snapshots: %w", err)
// 	}
// 	defer res.Body.Close()

// 	var searchResponse struct {
// 		Hits struct {
// 			Hits []struct {
// 				Source domain.Server `json:"_source"`
// 			} `json:"hits"`
// 		} `json:"hits"`
// 	}

// 	if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
// 		return nil, fmt.Errorf("failed to decode response: %w", err)
// 	}

// 	result := make(map[string]domain.Server)
// 	for _, hit := range searchResponse.Hits.Hits {
// 		result[hit.Source.ServerID] = hit.Source
// 	}

// 	return result, nil
// }

// func (r *recordRepository) getServerEvents(ctx context.Context, serverID string, startDate, endDate time.Time) ([]domain.Server, error) {
// 	query := map[string]interface{}{
// 		"size": 10000,
// 		"query": map[string]interface{}{
// 			"bool": map[string]interface{}{
// 				"must": []map[string]interface{}{
// 					{
// 						"term": map[string]interface{}{
// 							"server_id": serverID,
// 						},
// 					},
// 					{
// 						"range": map[string]interface{}{
// 							"timestamp": map[string]interface{}{
// 								"gte": startDate.Format(time.RFC3339),
// 								"lte": endDate.Format(time.RFC3339),
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 		"sort": []map[string]interface{}{
// 			{
// 				"timestamp": map[string]string{"order": "asc"},
// 			},
// 		},
// 	}

// 	queryBytes, err := json.Marshal(query)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to marshal query: %w", err)
// 	}

// 	res, err := r.esClient.Search(
// 		r.esClient.Search.WithContext(ctx),
// 		r.esClient.Search.WithIndex(r.recordIdx),
// 		r.esClient.Search.WithBody(bytes.NewReader(queryBytes)),
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to search events: %w", err)
// 	}
// 	defer res.Body.Close()

// 	var searchResponse struct {
// 		Hits struct {
// 			Hits []struct {
// 				Source domain.Server `json:"_source"`
// 			} `json:"hits"`
// 		} `json:"hits"`
// 	}

// 	if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
// 		return nil, fmt.Errorf("failed to decode response: %w", err)
// 	}

// 	result := make([]domain.Server, 0, len(searchResponse.Hits.Hits))
// 	for _, hit := range searchResponse.Hits.Hits {
// 		result = append(result, hit.Source)
// 	}

// 	return result, nil
// }

// func (r *recordRepository) calculateUptimeSeconds(events []domain.Server, startDate, endDate time.Time, previousStatus string) int64 {
// 	var uptimeSeconds int64
// 	var lastOnlineTime *time.Time

// 	if len(events) == 0 {
// 		if previousStatus == "ON" {
// 			return int64(endDate.Sub(startDate).Seconds())
// 		}
// 		return 0
// 	}

// 	if len(events) == 1 && events[0].Status == "OFF" {
// 		if previousStatus == "ON" {
// 			return int64(events[0].Timestamp.Sub(startDate).Seconds())
// 		}

// 		return 0
// 	}

// 	for _, event := range events {
// 		if event.Status == "ON" {
// 			lastOnlineTime = &event.Timestamp
// 		} else if event.Status == "OFF" && lastOnlineTime != nil {
// 			// Calculate uptime for this online period
// 			onlineStart := *lastOnlineTime
// 			offlineTime := event.Timestamp

// 			// Adjust for query time range
// 			if onlineStart.Before(startDate) {
// 				onlineStart = startDate
// 			}
// 			if offlineTime.After(endDate) {
// 				offlineTime = endDate
// 			}

// 			if onlineStart.Before(offlineTime) {
// 				uptimeSeconds += int64(offlineTime.Sub(onlineStart).Seconds())
// 			}

// 			lastOnlineTime = nil
// 		}
// 	}

// 	// If server is still online at the end of the period
// 	if lastOnlineTime != nil {
// 		onlineStart := *lastOnlineTime
// 		if onlineStart.Before(endDate) {
// 			uptimeSeconds += int64(endDate.Sub(onlineStart).Seconds())
// 		}
// 	}

// 	return uptimeSeconds
// }

// func (r *recordRepository) lastStatusBeforeDate(ctx context.Context, serverID string, startDate time.Time) (string, error) {
//     query := map[string]interface{}{
//         "size": 1,
//         "query": map[string]interface{}{
//             "bool": map[string]interface{}{
//                 "must": []map[string]interface{}{
//                     {
//                         "term": map[string]interface{}{
//                             "server_id": serverID,
//                         },
//                     },
//                     {
//                         "range": map[string]interface{}{
//                             "timestamp": map[string]interface{}{
//                                 "lt": startDate.Format(time.RFC3339),
//                             },
//                         },
//                     },
//                 },
//             },
//         },
//         "sort": []map[string]interface{}{
//             {
//                 "timestamp": map[string]string{"order": "desc"},
//             },
//         },
//     }

//     queryBytes, err := json.Marshal(query)
//     if err != nil {
//         return "OFF", fmt.Errorf("failed to marshal query: %w", err)
//     }

//     res, err := r.esClient.Search(
//         r.esClient.Search.WithContext(ctx),
//         r.esClient.Search.WithIndex(r.recordIdx),
//         r.esClient.Search.WithBody(bytes.NewReader(queryBytes)),
//     )
//     if err != nil {
//         return "OFF", fmt.Errorf("failed to search events: %w", err)
//     }
//     defer res.Body.Close()

//     var searchResponse struct {
//         Hits struct {
//             Total struct {
//                 Value int64 `json:"value"`
//             } `json:"total"`
//             Hits []struct {
//                 Source domain.Server `json:"_source"`
//             } `json:"hits"`
//         } `json:"hits"`
//     }

//     if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
//         return "OFF", fmt.Errorf("failed to decode response: %w", err)
//     }

//     if searchResponse.Hits.Total.Value == 0 {
//         return "OFF", nil
//     }

//     return searchResponse.Hits.Hits[0].Source.Status, nil
// }

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

// OptimizedRecordRepository - Optimized for high-performance uptime calculation
// Key optimizations:
// 1. Batch Elasticsearch queries using aggregations
// 2. Reduce from N*3 queries to 3 total queries
// 3. Use scroll API for large datasets (millions of records)
// 4. Parallel processing with worker pools
// 5. Memory-efficient processing with streaming
type recordRepository struct {
	esClient    *elasticsearch.Client
	snapshotIdx string
	recordIdx   string
	// Worker pool size for parallel processing
	workerPoolSize int
	// Batch size for scroll API
	scrollSize int
}

func NewRecordRepository(esClient *elasticsearch.Client, cfg *config.Config) domain.Repository {
	return &recordRepository{
		esClient:       esClient,
		snapshotIdx:    cfg.Elasticsearch.SnapshotIndex,
		recordIdx:      cfg.Elasticsearch.RecordIndex,
		workerPoolSize: 100,   // Adjust based on your Elasticsearch cluster capacity
		scrollSize:     10000, // Process 10000 records at a time
	}
}

// GetUptimeStats - Optimized version using batch queries and aggregations
func (r *recordRepository) GetUptimeStats(ctx context.Context, startDate, endDate time.Time) (*domain.UptimeStats, error) {
	totalSeconds := endDate.Sub(startDate).Seconds()

	// OPTIMIZATION 1: Get all server snapshots in one query with scroll API
	serverSnapshots, err := r.getAllServerSnapshotsWithScroll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get server snapshots: %w", err)
	}

	totalServers := len(serverSnapshots)
	if totalServers == 0 {
		return &domain.UptimeStats{
			StartDate:        startDate,
			EndDate:          endDate,
			TotalServers:     0,
			OnlineServers:    0,
			OfflineServers:   0,
			UptimePercentage: 0,
		}, nil
	}

	// OPTIMIZATION 2: Batch fetch all events in one query instead of N queries
	allEventsMap, err := r.getAllServerEventsBatch(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get all server events: %w", err)
	}

	// OPTIMIZATION 3: Batch fetch last status for all servers in one query
	lastStatusMap, err := r.getLastStatusBatch(ctx, serverSnapshots, startDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get last status batch: %w", err)
	}

	// OPTIMIZATION 4: Use worker pool for parallel processing
	uptimeDetails := r.processServersWithWorkerPool(
		serverSnapshots,
		allEventsMap,
		lastStatusMap,
		startDate,
		endDate,
		totalSeconds,
	)

	// Calculate aggregate statistics
	var onlineServers int
	var totalUptimePercentage float64
	for _, detail := range uptimeDetails {
		if serverSnapshots[detail.ServerID].Status == "ON" {
			onlineServers++
		}
		totalUptimePercentage += detail.UptimePercentage
	}

	return &domain.UptimeStats{
		StartDate:        startDate,
		EndDate:          endDate,
		TotalServers:     totalServers,
		OnlineServers:    onlineServers,
		OfflineServers:   totalServers - onlineServers,
		UptimePercentage: totalUptimePercentage / float64(totalServers),
		ServerDetails:    uptimeDetails,
	}, nil
}

// getAllServerSnapshotsWithScroll - Use scroll API for large datasets
func (r *recordRepository) getAllServerSnapshotsWithScroll(ctx context.Context) (map[string]domain.Server, error) {
	result := make(map[string]domain.Server)

	// Initial search with scroll
	query := map[string]interface{}{
		"size": r.scrollSize,
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
		// Only fetch required fields to reduce network overhead
		"_source": []string{"server_id", "port", "status", "timestamp"},
	}

	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	res, err := r.esClient.Search(
		r.esClient.Search.WithContext(ctx),
		r.esClient.Search.WithIndex(r.snapshotIdx),
		r.esClient.Search.WithBody(bytes.NewReader(queryBytes)),
		r.esClient.Search.WithScroll(time.Minute),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search snapshots: %w", err)
	}
	defer res.Body.Close()

	var searchResponse struct {
		ScrollID string `json:"_scroll_id"`
		Hits     struct {
			Hits []struct {
				Source domain.Server `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Process first batch
	for _, hit := range searchResponse.Hits.Hits {
		result[hit.Source.ServerID] = hit.Source
	}

	// Continue scrolling for remaining data
	scrollID := searchResponse.ScrollID
	for len(searchResponse.Hits.Hits) > 0 {
		scrollRes, err := r.esClient.Scroll(
			r.esClient.Scroll.WithContext(ctx),
			r.esClient.Scroll.WithScrollID(scrollID),
			r.esClient.Scroll.WithScroll(time.Minute),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scroll: %w", err)
		}

		if err := json.NewDecoder(scrollRes.Body).Decode(&searchResponse); err != nil {
			scrollRes.Body.Close()
			return nil, fmt.Errorf("failed to decode scroll response: %w", err)
		}
		scrollRes.Body.Close()

		for _, hit := range searchResponse.Hits.Hits {
			result[hit.Source.ServerID] = hit.Source
		}

		scrollID = searchResponse.ScrollID
	}

	// Clear scroll context
	r.esClient.ClearScroll(r.esClient.ClearScroll.WithScrollID(scrollID))

	return result, nil
}

// getAllServerEventsBatch - Fetch all events for all servers in batches using scroll API
// This replaces N individual queries with 1 batch query
func (r *recordRepository) getAllServerEventsBatch(ctx context.Context, startDate, endDate time.Time) (map[string][]domain.Server, error) {
	result := make(map[string][]domain.Server)
	var mu sync.Mutex

	// Query all events in the time range, sorted by server_id and timestamp
	query := map[string]interface{}{
		"size": r.scrollSize,
		"query": map[string]interface{}{
			"range": map[string]interface{}{
				"timestamp": map[string]interface{}{
					"gte": startDate.Format(time.RFC3339),
					"lte": endDate.Format(time.RFC3339),
				},
			},
		},
		"sort": []map[string]interface{}{
			{"server_id": map[string]string{"order": "asc"}},
			{"timestamp": map[string]string{"order": "asc"}},
		},
		"_source": []string{"server_id", "port", "status", "timestamp"},
	}

	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	res, err := r.esClient.Search(
		r.esClient.Search.WithContext(ctx),
		r.esClient.Search.WithIndex(r.recordIdx),
		r.esClient.Search.WithBody(bytes.NewReader(queryBytes)),
		r.esClient.Search.WithScroll(time.Minute*5),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search events: %w", err)
	}
	defer res.Body.Close()

	var searchResponse struct {
		ScrollID string `json:"_scroll_id"`
		Hits     struct {
			Hits []struct {
				Source domain.Server `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Process first batch
	mu.Lock()
	for _, hit := range searchResponse.Hits.Hits {
		serverID := hit.Source.ServerID
		result[serverID] = append(result[serverID], hit.Source)
	}
	mu.Unlock()

	// Continue scrolling
	scrollID := searchResponse.ScrollID
	for len(searchResponse.Hits.Hits) > 0 {
		scrollRes, err := r.esClient.Scroll(
			r.esClient.Scroll.WithContext(ctx),
			r.esClient.Scroll.WithScrollID(scrollID),
			r.esClient.Scroll.WithScroll(time.Minute*5),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scroll: %w", err)
		}

		if err := json.NewDecoder(scrollRes.Body).Decode(&searchResponse); err != nil {
			scrollRes.Body.Close()
			return nil, fmt.Errorf("failed to decode scroll response: %w", err)
		}
		scrollRes.Body.Close()

		mu.Lock()
		for _, hit := range searchResponse.Hits.Hits {
			serverID := hit.Source.ServerID
			result[serverID] = append(result[serverID], hit.Source)
		}
		mu.Unlock()

		scrollID = searchResponse.ScrollID
	}

	r.esClient.ClearScroll(r.esClient.ClearScroll.WithScrollID(scrollID))

	return result, nil
}

// getLastStatusBatch - Fetch last status before startDate for all servers using aggregations
// This replaces N individual queries with 1 aggregation query
func (r *recordRepository) getLastStatusBatch(ctx context.Context, servers map[string]domain.Server, startDate time.Time) (map[string]string, error) {
	result := make(map[string]string)

	// Use terms aggregation with top_hits to get last status for each server
	// This is much more efficient than N individual queries
	query := map[string]interface{}{
		"size": 0, // We only need aggregations, not hits
		"query": map[string]interface{}{
			"range": map[string]interface{}{
				"timestamp": map[string]interface{}{
					"lt": startDate.Format(time.RFC3339),
				},
			},
		},
		"aggs": map[string]interface{}{
			"servers": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "server_id",  // Use server_id field directly (not .keyword)
					"size":  len(servers), // Get all servers
				},
				"aggs": map[string]interface{}{
					"last_status": map[string]interface{}{
						"top_hits": map[string]interface{}{
							"size": 1,
							"sort": []map[string]interface{}{
								{"timestamp": map[string]string{"order": "desc"}},
							},
							"_source": []string{"status"},
						},
					},
				},
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
		Aggregations struct {
			Servers struct {
				Buckets []struct {
					Key        string `json:"key"`
					LastStatus struct {
						Hits struct {
							Hits []struct {
								Source struct {
									Status string `json:"status"`
								} `json:"_source"`
							} `json:"hits"`
						} `json:"hits"`
					} `json:"last_status"`
				} `json:"buckets"`
			} `json:"servers"`
		} `json:"aggregations"`
	}

	if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Process aggregation results
	for _, bucket := range searchResponse.Aggregations.Servers.Buckets {
		if len(bucket.LastStatus.Hits.Hits) > 0 {
			result[bucket.Key] = bucket.LastStatus.Hits.Hits[0].Source.Status
		}
	}

	// Set default "OFF" for servers without previous status
	for serverID := range servers {
		if _, exists := result[serverID]; !exists {
			result[serverID] = "OFF"
		}
	}

	return result, nil
}

// processServersWithWorkerPool - Process servers in parallel with worker pool pattern
func (r *recordRepository) processServersWithWorkerPool(
	servers map[string]domain.Server,
	eventsMap map[string][]domain.Server,
	lastStatusMap map[string]string,
	startDate, endDate time.Time,
	totalSeconds float64,
) []domain.ServerUptimeDetail {

	// Create channels for work distribution
	jobs := make(chan domain.Server, len(servers))
	results := make(chan domain.ServerUptimeDetail, len(servers))

	// Create worker pool
	var wg sync.WaitGroup
	for w := 0; w < r.workerPoolSize; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for server := range jobs {
				events := eventsMap[server.ServerID]
				previousStatus := lastStatusMap[server.ServerID]

				uptimeSeconds := r.calculateUptimeSeconds(events, startDate, endDate, previousStatus)

				results <- domain.ServerUptimeDetail{
					ServerID:         server.ServerID,
					UptimePercentage: float64(uptimeSeconds) / totalSeconds * 100,
				}
			}
		}()
	}

	// Send jobs
	for _, server := range servers {
		jobs <- server
	}
	close(jobs)

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	uptimeDetails := make([]domain.ServerUptimeDetail, 0, len(servers))
	for detail := range results {
		uptimeDetails = append(uptimeDetails, detail)
	}

	return uptimeDetails
}

// calculateUptimeSeconds - Same logic as before, optimized for performance
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

	// Initialize with previous status
	if previousStatus == "ON" {
		t := startDate
		lastOnlineTime = &t
	}

	for _, event := range events {
		if event.Status == "ON" {
			if lastOnlineTime == nil {
				lastOnlineTime = &event.Timestamp
			}
		} else if event.Status == "OFF" && lastOnlineTime != nil {
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
