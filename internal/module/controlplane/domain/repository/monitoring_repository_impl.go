package repository

import (
	"context"
	"time"

	"github.com/zhinea/sylix/internal/module/controlplane/entity"
	"gorm.io/gorm"
)

type MonitoringRepositoryImpl struct {
	db *gorm.DB
}

func NewMonitoringRepository(db *gorm.DB) MonitoringRepository {
	return &MonitoringRepositoryImpl{
		db: db,
	}
}

func (r *MonitoringRepositoryImpl) SavePing(ctx context.Context, ping *entity.ServerPing) error {
	return r.db.WithContext(ctx).Create(ping).Error
}

func (r *MonitoringRepositoryImpl) SaveStat(ctx context.Context, stat *entity.ServerStat) error {
	return r.db.WithContext(ctx).Create(stat).Error
}

func (r *MonitoringRepositoryImpl) SaveAccident(ctx context.Context, accident *entity.ServerAccident) error {
	return r.db.WithContext(ctx).Create(accident).Error
}

func (r *MonitoringRepositoryImpl) GetPingsByServerID(ctx context.Context, serverID string, since time.Time) ([]*entity.ServerPing, error) {
	var pings []*entity.ServerPing
	err := r.db.WithContext(ctx).Where("server_id = ? AND created_at >= ?", serverID, since).Find(&pings).Error
	return pings, err
}

func (r *MonitoringRepositoryImpl) GetRecentPings(ctx context.Context, serverID string, limit int) ([]*entity.ServerPing, error) {
	var pings []*entity.ServerPing
	err := r.db.WithContext(ctx).Where("server_id = ?", serverID).Order("created_at desc").Limit(limit).Find(&pings).Error
	return pings, err
}

func (r *MonitoringRepositoryImpl) GetStatsByServerID(ctx context.Context, serverID string, limit int) ([]*entity.ServerStat, error) {
	var stats []*entity.ServerStat
	err := r.db.WithContext(ctx).Where("server_id = ?", serverID).Order("created_at desc").Limit(limit).Find(&stats).Error
	return stats, err
}

func (r *MonitoringRepositoryImpl) GetAccidents(ctx context.Context, serverID string, startDate, endDate *time.Time, resolved *bool, offset, limit int) ([]*entity.ServerAccident, int64, error) {
	var accidents []*entity.ServerAccident
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.ServerAccident{})

	if serverID != "" {
		query = query.Where("server_id = ?", serverID)
	}
	if startDate != nil {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", endDate)
	}
	if resolved != nil {
		query = query.Where("resolved = ?", resolved)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at desc").Offset(offset).Limit(limit).Find(&accidents).Error
	return accidents, total, err
}

func (r *MonitoringRepositoryImpl) DeleteOldPings(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).Where("created_at < ?", before).Delete(&entity.ServerPing{}).Error
}

func (r *MonitoringRepositoryImpl) DeleteAccident(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.ServerAccident{}, "id = ?", id).Error
}

func (r *MonitoringRepositoryImpl) BatchDeleteAccidents(ctx context.Context, ids []string) error {
	return r.db.WithContext(ctx).Delete(&entity.ServerAccident{}, "id IN ?", ids).Error
}
