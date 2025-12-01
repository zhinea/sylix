package repository

import (
	"context"
	"time"

	"github.com/zhinea/sylix/internal/module/controlplane/entity"
)

type MonitoringRepository interface {
	SavePing(ctx context.Context, ping *entity.ServerPing) error
	SaveStat(ctx context.Context, stat *entity.ServerStat) error
	GetPingsByServerID(ctx context.Context, serverID string, since time.Time) ([]*entity.ServerPing, error)
	GetRecentPings(ctx context.Context, serverID string, limit int) ([]*entity.ServerPing, error)
	GetStatsByServerID(ctx context.Context, serverID string, limit int) ([]*entity.ServerStat, error)
	DeleteOldPings(ctx context.Context, before time.Time) error
}
