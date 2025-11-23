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

func (r *MonitoringRepositoryImpl) GetStatsByServerID(ctx context.Context, serverID string, limit int) ([]*entity.ServerStat, error) {
	var stats []*entity.ServerStat
	err := r.db.WithContext(ctx).Where("server_id = ?", serverID).Order("created_at desc").Limit(limit).Find(&stats).Error
	return stats, err
}

func (r *MonitoringRepositoryImpl) GetAccidentsByServerID(ctx context.Context, serverID string, limit int) ([]*entity.ServerAccident, error) {
	var accidents []*entity.ServerAccident
	err := r.db.WithContext(ctx).Where("server_id = ?", serverID).Order("created_at desc").Limit(limit).Find(&accidents).Error
	return accidents, err
}

func (r *MonitoringRepositoryImpl) DeleteOldPings(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).Where("created_at < ?", before).Delete(&entity.ServerPing{}).Error
}
