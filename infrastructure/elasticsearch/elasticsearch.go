package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/lits-06/vcs-sms/config"
	"github.com/lits-06/vcs-sms/services/server"
)

type SearchResult struct {
	Aggregations struct {
		Servers struct {
			Buckets []struct {
				Key          string `json:"key"`
				DocCount     int    `json:"doc_count"`
				LatestStatus struct {
					Hits struct {
						Hits []struct {
							Source struct {
								Status    string    `json:"status"`
								Timestamp time.Time `json:"timestamp"`
							} `json:"_source"`
						} `json:"hits"`
					} `json:"hits"`
				} `json:"latest_status"`
				OnlineIntervals struct {
					TotalOnlineTime struct {
						Value float64 `json:"value"`
					} `json:"total_online_time"`
				} `json:"online_intervals"`
			} `json:"buckets"`
		} `json:"servers"`
	} `json:"aggregations"`
}

func NewElasticsearchClient(cfg *config.ElasticsearchConfig) (*elasticsearch.Client, error) {
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{cfg.GetElasticsearchURL()},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	// Test connection
	res, err := es.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("Elasticsearch connection error: %s", res.String())
	}

	return es, nil
}

type elasticRepository struct {
	client *elasticsearch.Client
}

func NewRecordRepository(client *elasticsearch.Client) server.RecordRepository {
	repo := &elasticRepository{
		client: client,
	}

	repo.createIndexTemplate()

	return repo
}

func (r *elasticRepository) createIndexTemplate() {
	indexTemplate := map[string]interface{}{
		"index_patterns": []string{"server-status-*"},
		"template": map[string]interface{}{
			"mappings": map[string]interface{}{
				"properties": map[string]interface{}{
					"server_id": map[string]interface{}{
						"type": "keyword",
					},
					"status": map[string]interface{}{
						"type": "keyword",
					},
					"timestamp": map[string]interface{}{
						"type": "date",
					},
					"interval": map[string]interface{}{
						"type": "long",
					},
				},
			},
		},
	}

	templateJSON, _ := json.Marshal(indexTemplate)

	r.client.Indices.PutIndexTemplate(
		"server-status-template",
		bytes.NewReader(templateJSON),
	)
}

func (r *elasticRepository) CreateBatch(ctx context.Context, records []*server.StatusRecord) error {
	if len(records) == 0 {
		return nil // Nothing to do
	}

	var buf bytes.Buffer
	for _, record := range records {
		indexName := r.getIndexName(record.Timestamp)

		action := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": indexName,
			},
		}

		actionJSON, _ := json.Marshal(action)
		buf.Write(actionJSON)
		buf.WriteByte('\n')

		recordJSON, _ := json.Marshal(record)
		buf.Write(recordJSON)
		buf.WriteByte('\n')
	}

	res, err := r.client.Bulk(bytes.NewReader(buf.Bytes()), r.client.Bulk.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to bulk create records: %w", err)
	}

	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("Elasticsearch bulk create error: %s", res.String())
	}

	return nil
}

func (r *elasticRepository) GetUptimeStats(ctx context.Context, req *server.UptimeRequest) (*server.UptimeStats, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"range": map[string]interface{}{
				"timestamp": map[string]interface{}{
					"gte": req.StartDate.Format(time.RFC3339),
					"lte": req.EndDate.Format(time.RFC3339),
				},
			},
		},
		"aggs": map[string]interface{}{
			"servers": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "server_id",
					"size":  10000, // Giả sử tối đa 10k servers
				},
				"aggs": map[string]interface{}{
					// Lấy status record cuối cùng của mỗi server
					"latest_status": map[string]interface{}{
						"top_hits": map[string]interface{}{
							"sort": []map[string]interface{}{
								{
									"timestamp": map[string]interface{}{
										"order": "desc",
									},
								},
							},
							"size":    1, // Chỉ lấy record mới nhất
							"_source": []string{"status", "timestamp"},
						},
					},
					"online_intervals": map[string]interface{}{
						"filter": map[string]interface{}{
							"term": map[string]interface{}{
								"status": "ON",
							},
						},
						"aggs": map[string]interface{}{
							"total_online_time": map[string]interface{}{
								"sum": map[string]interface{}{
									"field": "interval",
								},
							},
						},
					},
				},
			},
		},
		"size": 0, // Chỉ cần aggregation, không cần documents
	}

	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	indices := r.getIndicesForTimeRange(req.StartDate, req.EndDate)
	indexPattern := strings.Join(indices, ",")

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex(indexPattern),
		r.client.Search.WithBody(bytes.NewReader(queryJSON)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("Elasticsearch search error: %s", res.String())
	}

	var result SearchResult
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	totalServers := len(result.Aggregations.Servers.Buckets)
	onlineServers := 0
	offlineServers := 0
	totalOnlineTime := 0.0

	for _, bucket := range result.Aggregations.Servers.Buckets {
		totalOnlineTime += bucket.OnlineIntervals.TotalOnlineTime.Value

		if len(bucket.LatestStatus.Hits.Hits) > 0 {
			status := bucket.LatestStatus.Hits.Hits[0].Source.Status
			if status == "ON" {
				onlineServers++
			} else {
				offlineServers++
			}
		} else {
			offlineServers++
		}
	}

	uptimePercentage := totalOnlineTime / (float64(totalServers) * float64(req.EndDate.Sub(req.StartDate).Seconds())) * 100

	return &server.UptimeStats{
		TotalServers:     totalServers,
		OnlineServers:    onlineServers,
		OfflineServers:   offlineServers,
		UptimePercentage: uptimePercentage,
	}, nil
}

func (r *elasticRepository) getIndexName(timestamp time.Time) string {
	return fmt.Sprintf("server-status-%s", timestamp.Format("2006.01.02"))
}

func (r *elasticRepository) getIndicesForTimeRange(startTime, endTime time.Time) []string {
	var indices []string
	current := startTime

	for current.Before(endTime) || current.Equal(endTime) {
		indexName := fmt.Sprintf("server-status-%s", current.Format("2006.01.02"))
		indices = append(indices, indexName)
		current = current.AddDate(0, 0, 1) // Thêm 1 ngày
	}

	if len(indices) == 0 {
		// Fallback: sử dụng pattern để search tất cả indices
		indices = []string{"server-status-*"}
	}

	return indices
}
