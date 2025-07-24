package server

import (
	"context"
	"fmt"
	"time"

	"github.com/lits-06/vcs-sms/pkg/logger"
)

type UptimeService struct {
	consumer MessageConsumer
	repo     Record
	logger   logger.Logger
}

func NewUptimeConsumerService(
	consumer MessageConsumer,
	repo Record,
	logger logger.Logger,
) *UptimeConsumerService {
	return &UptimeConsumerService{
		consumer: consumer,
		repo:     repo,
		logger:   logger.With("service", "uptime_consumer"),
	}
}

// StartConsuming starts consuming server status messages and storing uptime records
func (s *UptimeService) StartConsuming(ctx context.Context) error {
	return s.consumer.ConsumeServerStatus(ctx, s.handleServerStatus)
}

// handleServerStatus processes incoming server status messages
func (s *UptimeService) handleServerStatus(rc *StatusRecord) error {
	// Convert to status record
	// record := StatusRecord{
	// 	ServerID:  msg.ServerID,
	// 	Status:    msg.Status,
	// 	Timestamp: msg.Timestamp,
	// }

	// Store in Elasticsearch
	if err := s.storeUptimeRecord(rc); err != nil {
		s.logger.Error("Failed to store uptime record",
			"error", err,
			"server_id", rc.ServerID,
		)
		return err
	}

	return nil
}

// storeUptimeRecord stores uptime record in Elasticsearch
func (s *UptimeConsumerService) storeUptimeRecord(rc *UptimeRecord) error {
	ctx := context.Background()

	err := s.repo.Create(ctx, rc)
	if err != nil {
		return fmt.Errorf("failed to index uptime record: %w", err)
	}

	return nil
}

// CalculateGlobalUptime calculates uptime percentage for all servers in time range
func (s *UptimeConsumerService) CalculateGlobalUptime(ctx context.Context, from, to time.Time) (*UptimeStats, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"range": map[string]interface{}{
				"timestamp": map[string]interface{}{
					"gte": from.Format(time.RFC3339),
					"lte": to.Format(time.RFC3339),
				},
			},
		},
		"aggs": map[string]interface{}{
			"servers": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "server_id.keyword",
					"size":  10000, // Max servers to analyze
				},
				"aggs": map[string]interface{}{
					"status_count": map[string]interface{}{
						"terms": map[string]interface{}{
							"field": "status.keyword",
						},
					},
				},
			},
			"total_status": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "status.keyword",
				},
			},
		},
		"size": 0, // Chỉ lấy aggregation, không lấy documents
	}

	// Search across multiple indices
	indexPattern := fmt.Sprintf("server-uptime-%s*", from.Format("2006-01"))
	results, err := s.esClient.Search(ctx, indexPattern, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search uptime records: %w", err)
	}

	return s.parseUptimeResults(results, from, to)
}

// CalculateServerUptime calculates uptime for a specific server
func (s *UptimeConsumerService) CalculateServerUptime(ctx context.Context, serverID string, from, to time.Time) (*UptimeStats, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"server_id.keyword": serverID,
						},
					},
					{
						"range": map[string]interface{}{
							"timestamp": map[string]interface{}{
								"gte": from.Format(time.RFC3339),
								"lte": to.Format(time.RFC3339),
							},
						},
					},
				},
			},
		},
		"aggs": map[string]interface{}{
			"status_count": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "status.keyword",
				},
			},
		},
		"size": 0,
	}

	indexPattern := fmt.Sprintf("server-uptime-%s*", from.Format("2006-01"))
	results, err := s.esClient.Search(ctx, indexPattern, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search server uptime: %w", err)
	}

	return s.parseServerUptimeResults(results, serverID, from, to)
}

// parseUptimeResults parses Elasticsearch results for global uptime
func (s *UptimeConsumerService) parseUptimeResults(results *elasticsearch.SearchResult, from, to time.Time) (*UptimeStats, error) {
	stats := &UptimeStats{
		PeriodStart: from,
		PeriodEnd:   to,
	}

	// Parse aggregations
	aggs, ok := results.Aggregations["servers"].(map[string]interface{})
	if !ok {
		return stats, nil
	}

	buckets, ok := aggs["buckets"].([]interface{})
	if !ok {
		return stats, nil
	}

	stats.TotalServers = len(buckets)

	for _, bucket := range buckets {
		bucketMap := bucket.(map[string]interface{})

		// Get status counts for this server
		statusAgg := bucketMap["status_count"].(map[string]interface{})
		statusBuckets := statusAgg["buckets"].([]interface{})

		var onlineCount, offlineCount int
		for _, statusBucket := range statusBuckets {
			statusBucketMap := statusBucket.(map[string]interface{})
			status := statusBucketMap["key"].(string)
			count := int(statusBucketMap["doc_count"].(float64))

			if status == "ON" || status == "online" {
				onlineCount += count
			} else {
				offlineCount += count
			}
		}

		stats.OnlineChecks += onlineCount
		stats.OfflineChecks += offlineCount

		// Count server as online if it has more online checks than offline
		if onlineCount > offlineCount {
			stats.OnlineServers++
		} else {
			stats.OfflineServers++
		}
	}

	stats.TotalChecks = stats.OnlineChecks + stats.OfflineChecks

	if stats.TotalChecks > 0 {
		stats.UptimePercentage = float64(stats.OnlineChecks) / float64(stats.TotalChecks) * 100
	}

	return stats, nil
}

// parseServerUptimeResults parses results for specific server uptime
func (s *UptimeConsumerService) parseServerUptimeResults(results *elasticsearch.SearchResult, serverID string, from, to time.Time) (*UptimeStats, error) {
	stats := &UptimeStats{
		TotalServers: 1,
		PeriodStart:  from,
		PeriodEnd:    to,
	}

	aggs, ok := results.Aggregations["status_count"].(map[string]interface{})
	if !ok {
		return stats, nil
	}

	buckets, ok := aggs["buckets"].([]interface{})
	if !ok {
		return stats, nil
	}

	for _, bucket := range buckets {
		bucketMap := bucket.(map[string]interface{})
		status := bucketMap["key"].(string)
		count := int(bucketMap["doc_count"].(float64))

		if status == "ON" || status == "online" {
			stats.OnlineChecks = count
			stats.OnlineServers = 1
		} else {
			stats.OfflineChecks = count
			stats.OfflineServers = 1
		}
	}

	stats.TotalChecks = stats.OnlineChecks + stats.OfflineChecks

	if stats.TotalChecks > 0 {
		stats.UptimePercentage = float64(stats.OnlineChecks) / float64(stats.TotalChecks) * 100
	}

	return stats, nil
}
