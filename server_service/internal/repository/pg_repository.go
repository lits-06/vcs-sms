package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/lits-06/vcs-sms/pkg/tracing"
	"github.com/lits-06/vcs-sms/server_service/internal/domain"
	"github.com/opentracing/opentracing-go"

	"gorm.io/gorm"
)

type serverRepository struct {
	db *gorm.DB
}

func NewServerRepository(db *gorm.DB) domain.Repository {
	return &serverRepository{
		db: db,
	}
}

func (r *serverRepository) Create(ctx context.Context, srv *domain.Server) (*domain.Server, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "serverRepository.Create")
	defer span.Finish()

	if err := r.db.WithContext(ctx).Create(srv).Error; err != nil {
		return nil, tracing.TraceWithErr(span, fmt.Errorf("failed to create server: %w", err))
	}
	return srv, nil
}

func (r *serverRepository) CreateBatch(ctx context.Context, servers []domain.Server) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "serverRepository.CreateBatch")
	defer span.Finish()

	if err := r.db.WithContext(ctx).Create(&servers).Error; err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to create servers batch: %w", err))
	}
	return nil
}

func (r *serverRepository) GetByID(ctx context.Context, id string) (*domain.Server, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "serverRepository.GetByID")
	defer span.Finish()

	var srv domain.Server
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&srv).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Not found
		}
		return nil, tracing.TraceWithErr(span, fmt.Errorf("failed to get server by ID: %w", err))
	}
	return &srv, nil
}

func (r *serverRepository) GetByName(ctx context.Context, name string) (*domain.Server, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "serverRepository.GetByName")
	defer span.Finish()

	var srv domain.Server
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&srv).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, tracing.TraceWithErr(span, fmt.Errorf("failed to get server by name: %w", err))
	}
	return &srv, nil
}

func (r *serverRepository) Update(ctx context.Context, id, name, ipv4 string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "serverRepository.Update")
	defer span.Finish()

	data := make(map[string]interface{})
	if name != "" {
		data["name"] = name
	}

	if ipv4 != "" {
		data["ipv4"] = ipv4
	}

	result := r.db.WithContext(ctx).Model(&domain.Server{}).Where("id = ?", id).Updates(data)

	if result.Error != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to update server: %w", result.Error))
	}

	return nil
}

func (r *serverRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "serverRepository.UpdateStatus")
	defer span.Finish()

	result := r.db.WithContext(ctx).Model(&domain.Server{}).Where("id = ?", id).Update("status", status)

	if result.Error != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to update server status: %w", result.Error))
	}

	if result.RowsAffected == 0 {
		return tracing.TraceWithErr(span, fmt.Errorf("server with ID %s not found", id))
	}

	return nil
}

func (r *serverRepository) Delete(ctx context.Context, id string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "serverRepository.Delete")
	defer span.Finish()

	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.Server{})

	if result.Error != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to delete server: %w", result.Error))
	}

	if result.RowsAffected == 0 {
		return tracing.TraceWithErr(span, fmt.Errorf("server with ID %s not found", id))
	}

	return nil
}

func (r *serverRepository) List(ctx context.Context, name, status, ipv4 string, from, to int, sort, order string) (*[]domain.Server, int, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "serverRepository.List")
	defer span.Finish()

	var servers []domain.Server
	var total int64

	// Build base query with filters
	query := r.db.WithContext(ctx).Model(&domain.Server{})
	query = r.applyFilters(query, name, status, ipv4)

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, tracing.TraceWithErr(span, fmt.Errorf("failed to count servers: %w", err))
	}

	// Apply sorting
	query = r.applySorting(query, sort, order)

	// Apply pagination
	query = r.applyPagination(query, from, to)

	// Execute query
	if err := query.Find(&servers).Error; err != nil {
		return nil, 0, tracing.TraceWithErr(span, fmt.Errorf("failed to list servers: %w", err))
	}

	return &servers, int(total), nil
}

func (r *serverRepository) applyFilters(query *gorm.DB, name, status, ipv4 string) *gorm.DB {
	if name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if ipv4 != "" {
		query = query.Where("ipv4 = ?", ipv4)
	}

	return query
}

func (r *serverRepository) applySorting(query *gorm.DB, sort, order string) *gorm.DB {
	fielddf := "created_at" // Default sort field
	if sort != "" {
		fielddf = sort
	}

	orderdf := "ASC" // Default order
	if strings.ToLower(order) == "desc" {
		orderdf = "DESC"
	}

	return query.Order(fmt.Sprintf("%s %s", fielddf, orderdf))
}

func (r *serverRepository) applyPagination(query *gorm.DB, from, to int) *gorm.DB {
	if to < from {
		return query
	}

	if from > 0 {
		query = query.Offset(from)
	}

	if to > 0 {
		limit := to
		if from > 0 {
			limit = to - from
		}
		query = query.Limit(limit)
	}

	return query
}

func (r *serverRepository) ExistsWithID(ctx context.Context, id string) (bool, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "serverRepository.ExistsWithID")
	defer span.Finish()

	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Server{}).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return false, tracing.TraceWithErr(span, fmt.Errorf("failed to check if server exists by ID: %w", err))
	}
	return count > 0, nil
}

func (r *serverRepository) ExistsWithName(ctx context.Context, name string) (bool, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "serverRepository.ExistsWithName")
	defer span.Finish()

	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Server{}).Where("name = ?", name).Count(&count).Error
	if err != nil {
		return false, tracing.TraceWithErr(span, fmt.Errorf("failed to check if server exists by name: %w", err))
	}
	return count > 0, nil
}
