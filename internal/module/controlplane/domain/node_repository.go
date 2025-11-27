package domain

import (
	"context"

	"github.com/zhinea/sylix/internal/module/controlplane/entity"
)

type NodeRepository interface {
	Create(ctx context.Context, node *entity.Node) error
	Get(ctx context.Context, id string) (*entity.Node, error)
	List(ctx context.Context, offset, limit int) ([]*entity.Node, int64, error)
	Update(ctx context.Context, node *entity.Node) error
	Delete(ctx context.Context, id string) error
	GetByServerID(ctx context.Context, serverID string) ([]*entity.Node, error)
}
