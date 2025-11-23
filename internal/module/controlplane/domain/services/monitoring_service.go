package services

import (
	"context"
	"time"

	"github.com/zhinea/sylix/internal/module/controlplane/domain/repository"
	"github.com/zhinea/sylix/internal/module/controlplane/entity"
)

type MonitoringService struct {
	repo repository.MonitoringRepository
}

func NewMonitoringService(repo repository.MonitoringRepository) *MonitoringService {
	return &MonitoringService{
		repo: repo,
	}
}

func (s *MonitoringService) GetStats(ctx context.Context, serverID string) ([]*entity.ServerStat, error) {
	return s.repo.GetStatsByServerID(ctx, serverID, 100) // Limit to last 100 stats
}

func (s *MonitoringService) GetRealtimeStats(ctx context.Context, serverID string, limit int) ([]*entity.ServerPing, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.GetRecentPings(ctx, serverID, limit)
}

func (s *MonitoringService) GetAccidents(ctx context.Context, serverID string, startDate, endDate *time.Time, resolved *bool, page, pageSize int) ([]*entity.ServerAccident, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	return s.repo.GetAccidents(ctx, serverID, startDate, endDate, resolved, offset, pageSize)
}

func (s *MonitoringService) DeleteAccident(ctx context.Context, id string) error {
	return s.repo.DeleteAccident(ctx, id)
}

func (s *MonitoringService) BatchDeleteAccidents(ctx context.Context, ids []string) error {
	return s.repo.BatchDeleteAccidents(ctx, ids)
}
