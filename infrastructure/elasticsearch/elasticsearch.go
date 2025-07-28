package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/esapi"
	"github.com/lits-06/vcs-sms/config"
	"github.com/lits-06/vcs-sms/usecases/server"
)

type searchResult struct {
	Aggregations struct {
		Servers struct {
			Buckets []struct {
				Key         string `json:"key"`
				DocCount    int    `json:"doc_count"`
				StatusStats struct {
					Buckets []struct {
						Key      string `json:"key"`
						DocCount int    `json:"doc_count"`
					} `json:"buckets"`
				} `json:"status_stats"`
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
	res, err := client.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("Elasticsearch connection error: %s", res.String())
	}

	return client, nil
}

type elasticRepository struct {
	client *elasticsearch.Client
}

func NewRecordRepository(client *elasticsearch.Client) server.RecordRepository {
	repo := &elasticRepository{
		client: client,
	}

	// Tạo index template khi khởi tạo
	if err := repo.createIndexTemplate(context.Background()); err != nil {
		fmt.Printf("Warning: Failed to create index template: %v\n", err)
	}

	return repo
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

// GetUptimeStats tính toán thống kê uptime cho tất cả server trong khoảng thời gian
func (r *elasticRepository) GetUptimeStats(ctx context.Context, from, to time.Time) (*server.UptimeStats, error) {
	indexPattern := r.getIndexPattern(from, to)

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
					"field": "server_id",
					"size":  10000, // Giới hạn 10000 servers
				},
				"aggs": map[string]interface{}{
					"status_stats": map[string]interface{}{
						"terms": map[string]interface{}{
							"field": "status",
						},
					},
				},
			},
		},
		"size": 0, // Chỉ lấy aggregations, không lấy documents
	}

	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req := esapi.SearchRequest{
		Index: []string{indexPattern},
		Body:  bytes.NewReader(queryJSON),
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search error: %s", res.String())
	}

	var searchResult searchResult

	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode search result: %w", err)
	}

	return r.calculateUptimeStats(searchResult), nil
}

// calculateUptimeStats tính toán thống kê từ kết quả Elasticsearch
func (r *elasticsearchRecordRepository) calculateUptimeStats(result searchResult) *server.UptimeStats {
	totalServers := len(result.Aggregations.Servers.Buckets)
	onlineServers := 0
	offlineServers := 0
	totalUptimePercentage := 0.0

	for _, serverBucket := range result.Aggregations.Servers.Buckets {
		onlineCount := 0
		totalCount := 0

		for _, statusBucket := range serverBucket.StatusStats.Buckets {
			totalCount += statusBucket.DocCount
			if statusBucket.Key == "ON" {
				onlineCount += statusBucket.DocCount
			}
		}

		if totalCount > 0 {
			serverUptimePercentage := float64(onlineCount) / float64(totalCount) * 100
			totalUptimePercentage += serverUptimePercentage

			// Xác định server hiện tại là online hay offline dựa trên majority
			if serverUptimePercentage >= 50 {
				onlineServers++
			} else {
				offlineServers++
			}
		} else {
			offlineServers++
		}
	}

	averageUptime := 0.0
	if totalServers > 0 {
		averageUptime = totalUptimePercentage / float64(totalServers)
	}

	return &server.UptimeStats{
		TotalServers:     totalServers,
		OnlineServers:    onlineServers,
		OfflineServers:   offlineServers,
		UptimePercentage: averageUptime,
	}
}

// getIndexName tạo tên index theo ngày
func (e *elasticRepository) getIndexName(timestamp time.Time) string {
	return fmt.Sprintf("server-record-%s", timestamp.Format("2006-01-02"))
}

// getIndexPattern tạo pattern để search trong nhiều index
func (e *elasticRepository) getIndexPattern(from, to time.Time) string {
	// Nếu cùng tháng, sử dụng pattern cụ thể
	if from.Year() == to.Year() && from.Month() == to.Month() {
		return fmt.Sprintf("server-record-%s*", from.Format("2006-01"))
	}
	// Nếu khác tháng, sử dụng pattern rộng hơn
	return "server-record-*"
}

// createIndexTemplate tạo index template để định nghĩa mapping
func (r *elasticsearchRecordRepository) createIndexTemplate(ctx context.Context) error {
	template := map[string]interface{}{
		"index_patterns": []string{"server-record-*"},
		"template": map[string]interface{}{
			"settings": map[string]interface{}{
				"number_of_shards":   1,
				"number_of_replicas": 0,
				"refresh_interval":   "5s",
			},
			"mappings": map[string]interface{}{
				"properties": map[string]interface{}{
					"server_id": map[string]interface{}{
						"type": "keyword",
					},
					"status": map[string]interface{}{
						"type": "keyword",
					},
					"timestamp": map[string]interface{}{
						"type":   "date",
						"format": "strict_date_optional_time",
					},
				},
			},
		},
	}

	templateJSON, err := json.Marshal(template)
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	req := esapi.IndicesPutIndexTemplateRequest{
		Name: "server-record-template",
		Body: bytes.NewReader(templateJSON),
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return fmt.Errorf("failed to create index template: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("Elasticsearch index template error: %s", res.String())
	}

	return nil
}
