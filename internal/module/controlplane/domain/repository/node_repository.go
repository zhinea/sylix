package repository

import (
	"context"

	"github.com/zhinea/sylix/internal/module/controlplane/entity"
)

type NodeRepository interface {
	Create(ctx context.Context, graph *entity.NodeGraph) error
	Get(ctx context.Context, id string) (*entity.NodeGraph, error)
	Update(ctx context.Context, graph *entity.NodeGraph) error
	Delete(ctx context.Context, id string) error
	// Add other methods if needed
}
