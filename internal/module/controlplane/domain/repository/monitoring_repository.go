package repository

import (
	"context"
	"time"

	"github.com/zhinea/sylix/internal/module/controlplane/entity"
)

type MonitoringRepository interface {
	SavePing(ctx context.Context, ping *entity.ServerPing) error
	SaveStat(ctx context.Context, stat *entity.ServerStat) error
	SaveAccident(ctx context.Context, accident *entity.ServerAccident) error

	GetPingsByServerID(ctx context.Context, serverID string, since time.Time) ([]*entity.ServerPing, error)
	GetStatsByServerID(ctx context.Context, serverID string, limit int) ([]*entity.ServerStat, error)
	GetAccidentsByServerID(ctx context.Context, serverID string, limit int) ([]*entity.ServerAccident, error)
	DeleteOldPings(ctx context.Context, before time.Time) error
}
