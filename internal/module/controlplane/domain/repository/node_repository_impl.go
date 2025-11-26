package repository

import (
	"context"

	"github.com/zhinea/sylix/internal/module/controlplane/entity"
	"gorm.io/gorm"
)

type NodeRepositoryImpl struct {
	db *gorm.DB
}

func NewNodeRepository(db *gorm.DB) NodeRepository {
	return &NodeRepositoryImpl{db: db}
}

func (r *NodeRepositoryImpl) Create(ctx context.Context, graph *entity.NodeGraph) error {
	return r.db.WithContext(ctx).Create(graph).Error
}

func (r *NodeRepositoryImpl) Get(ctx context.Context, id string) (*entity.NodeGraph, error) {
	var graph entity.NodeGraph
	if err := r.db.WithContext(ctx).First(&graph, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &graph, nil
}

func (r *NodeRepositoryImpl) Update(ctx context.Context, graph *entity.NodeGraph) error {
	return r.db.WithContext(ctx).Save(graph).Error
}

func (r *NodeRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.NodeGraph{}, "id = ?", id).Error
}
