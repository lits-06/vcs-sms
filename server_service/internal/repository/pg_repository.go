package repository

import (
	"context"
	"fmt"

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

func (r *serverRepository) Create(ctx context.Context, srv *domain.Server) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "serverRepository.Create")
	defer span.Finish()

	if err := r.db.WithContext(ctx).Create(srv).Error; err != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to create server: %w", err))
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

func (r *serverRepository) Update(ctx context.Context, srv *domain.Server) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "serverRepository.Update")
	defer span.Finish()

	data := make(map[string]interface{})
	if srv.Name != "" {
		data["name"] = srv.Name
	}
	if srv.Status != "" {
		data["status"] = srv.Status
	}
	if srv.IPv4 != "" {
		data["ipv4"] = srv.IPv4
	}

	result := r.db.WithContext(ctx).Model(srv).Where("id = ?", srv.ID).Updates(data)

	if result.Error != nil {
		return tracing.TraceWithErr(span, fmt.Errorf("failed to update server: %w", result.Error))
	}

	if result.RowsAffected == 0 {
		return tracing.TraceWithErr(span, fmt.Errorf("server with ID %s not found", srv.ID))
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

func (r *serverRepository) List(ctx context.Context, filter domain.ServerFilter, sort domain.ServerSort, pagination domain.ServerPagination) (*[]domain.Server, int, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "serverRepository.List")
	defer span.Finish()

	var servers []domain.Server
	var total int64

	// Build base query with filters
	query := r.db.WithContext(ctx).Model(&domain.Server{})
	query = r.applyFilters(query, filter)

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, tracing.TraceWithErr(span, fmt.Errorf("failed to count servers: %w", err))
	}

	// Apply sorting
	query = r.applySorting(query, sort)

	// Apply pagination
	query = r.applyPagination(query, pagination)

	// Execute query
	if err := query.Find(&servers).Error; err != nil {
		return nil, 0, tracing.TraceWithErr(span, fmt.Errorf("failed to list servers: %w", err))
	}

	return &servers, int(total), nil
}

func (r *serverRepository) applyFilters(query *gorm.DB, filter domain.ServerFilter) *gorm.DB {
	if filter.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Name+"%")
	}

	if filter.Status != "" {
		query = query.Where("status = ?", string(filter.Status))
	}

	if filter.IPv4 != "" {
		query = query.Where("ipv4 = ?", filter.IPv4)
	}

	return query
}

func (r *serverRepository) applySorting(query *gorm.DB, sort domain.ServerSort) *gorm.DB {
	field := "created_at" // Default sort field
	if sort.Sort != "" {
		field = sort.Sort
	}

	order := "ASC" // Default order
	if sort.Order == domain.SortDesc {
		order = "DESC"
	}

	return query.Order(fmt.Sprintf("%s %s", field, order))
}

func (r *serverRepository) applyPagination(query *gorm.DB, pagination domain.ServerPagination) *gorm.DB {
	if pagination.To < pagination.From {
		return query
	}

	if pagination.From > 0 {
		query = query.Offset(pagination.From)
	}

	if pagination.To > 0 {
		limit := pagination.To
		if pagination.From > 0 {
			limit = pagination.To - pagination.From
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
