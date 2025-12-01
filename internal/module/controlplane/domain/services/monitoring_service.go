package services

import (
	"context"

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
