package services

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"time"

// 	"go.uber.org/zap"

// 	"VCS-Checkpoint1/internal/domain"
// 	"VCS-Checkpoint1/internal/infrastructure/elasticsearch"
// )

// type UptimeService struct {
// 	es     *elasticsearch.Client
// 	logger *zap.Logger
// }

// func NewUptimeService(es *elasticsearch.Client, logger *zap.Logger) *UptimeService {
// 	return &UptimeService{
// 		es:     es,
// 		logger: logger,
// 	}
// }

// func (s *UptimeService) RecordServerStatus(ctx context.Context, serverID int64, status string) error {
// 	record := domain.UptimeRecord{
// 		ServerID:  serverID,
// 		Status:    status,
// 		Timestamp: time.Now(),
// 	}

// 	data, err := json.Marshal(record)
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal uptime record: %w", err)
// 	}

// 	indexName := fmt.Sprintf("server-uptime-%s", time.Now().Format("2006-01"))

// 	err = s.es.Index(ctx, indexName, string(data))
// 	if err != nil {
// 		s.logger.Error("Failed to index uptime record",
// 			zap.Error(err),
// 			zap.Int64("server_id", serverID),
// 			zap.String("status", status),
// 		)
// 		return fmt.Errorf("failed to index uptime record: %w", err)
// 	}

// 	return nil
// }

// func (s *UptimeService) CalculateUptime(ctx context.Context, serverID int64, from, to time.Time) (*domain.UptimeStats, error) {
// 	query := map[string]interface{}{
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
// 								"gte": from.Format(time.RFC3339),
// 								"lte": to.Format(time.RFC3339),
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 		"aggs": map[string]interface{}{
// 			"status_count": map[string]interface{}{
// 				"terms": map[string]interface{}{
// 					"field": "status.keyword",
// 				},
// 			},
// 		},
// 		"sort": []map[string]interface{}{
// 			{
// 				"timestamp": map[string]interface{}{
// 					"order": "asc",
// 				},
// 			},
// 		},
// 	}

// 	indexPattern := fmt.Sprintf("server-uptime-%s*", from.Format("2006-01"))
// 	results, err := s.es.Search(ctx, indexPattern, query)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to search uptime records: %w", err)
// 	}

// 	// Parse results and calculate uptime percentage
// 	var records []domain.UptimeRecord
// 	for _, hit := range results.Hits.Hits {
// 		var record domain.UptimeRecord
// 		if err := json.Unmarshal(hit.Source, &record); err != nil {
// 			continue
// 		}
// 		records = append(records, record)
// 	}

// 	return s.calculateUptimeFromRecords(records, from, to), nil
// }

// func (s *UptimeService) calculateUptimeFromRecords(records []domain.UptimeRecord, from, to time.Time) *domain.UptimeStats {
// 	if len(records) == 0 {
// 		return &domain.UptimeStats{
// 			UptimePercentage: 0,
// 			TotalChecks:      0,
// 			OnlineChecks:     0,
// 			OfflineChecks:    0,
// 		}
// 	}

// 	onlineCount := 0
// 	totalCount := len(records)

// 	for _, record := range records {
// 		if record.Status == "online" {
// 			onlineCount++
// 		}
// 	}

// 	uptimePercentage := float64(onlineCount) / float64(totalCount) * 100

// 	return &domain.UptimeStats{
// 		UptimePercentage: uptimePercentage,
// 		TotalChecks:      totalCount,
// 		OnlineChecks:     onlineCount,
// 		OfflineChecks:    totalCount - onlineCount,
// 	}
// }
