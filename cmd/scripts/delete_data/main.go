package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/lits-06/vcs-sms/config"
	"github.com/lits-06/vcs-sms/entity"
	"github.com/lits-06/vcs-sms/infrastructure/database"
	infraRedis "github.com/lits-06/vcs-sms/infrastructure/redis"
	"github.com/lits-06/vcs-sms/pkg/logger"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func deleteElasticsearchData() error {
	today := time.Now().Format("2006.01.02")
	url := "http://localhost:9200/server-status-" + today

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	// Nếu có auth: req.SetBasicAuth("elastic", "your-password")

	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to delete index: %s", resp.Status)
	}

	fmt.Println("✅ Đã xóa index Elasticsearch")
	return nil
}

func deletePostgresData(db *gorm.DB) error {
	db.Delete(&entity.Server{}, "1=1") // Xóa tất cả dữ liệu trong bảng Server
	fmt.Println("✅ Đã xóa dữ liệu trong bảng Server")
	return nil
}

func flushRedisData(rdb *redis.Client) error {
	ctx := context.Background()
	if err := rdb.FlushAll(ctx).Err(); err != nil {
		return err
	}
	fmt.Println("✅ Đã xóa toàn bộ dữ liệu Redis")
	return nil
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	appLogger, err := logger.NewZapLogger(&cfg.Logging)
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}

	db, err := database.NewGormConnection(&cfg.Database)
	if err != nil {
		appLogger.Fatal("Failed to connect to database", "error", err)
	}

	rdb, err := infraRedis.NewRedisClient(&cfg.Redis)
	if err != nil {
		appLogger.Fatal("Failed to connect to Redis", "error", err)
	}

	// esClient, err := elasticsearch.NewElasticsearchClient(&cfg.Elasticsearch)
	// if err != nil {
	// 	appLogger.Fatal("Failed to connect to Elasticsearch", "error", err)
	// }

	// Gọi các hàm xóa
	if err := deleteElasticsearchData(); err != nil {
		log.Println("❌ Lỗi xóa Elasticsearch:", err)
	}

	if err := deletePostgresData(db); err != nil {
		log.Println("❌ Lỗi xóa dữ liệu Postgres:", err)
	}

	if err := flushRedisData(rdb); err != nil {
		log.Println("❌ Lỗi xóa Redis:", err)
	}
}
