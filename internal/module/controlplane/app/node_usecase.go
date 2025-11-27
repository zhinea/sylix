package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zhinea/sylix/internal/module/controlplane/domain"
	"github.com/zhinea/sylix/internal/module/controlplane/entity"
)

type NodeUseCase struct {
	nodeRepo domain.NodeRepository
}

func NewNodeUseCase(nodeRepo domain.NodeRepository) *NodeUseCase {
	return &NodeUseCase{
		nodeRepo: nodeRepo,
	}
}

func (uc *NodeUseCase) Create(ctx context.Context, name, description string, nodeType entity.NodeType, priorityStartup int32, fields, imports, exports, serverID string) (*entity.Node, error) {
	node := &entity.Node{
		ID:              uuid.New().String(),
		Name:            name,
		Description:     description,
		Type:            nodeType,
		PriorityStartup: priorityStartup,
		Fields:          fields,
		Imports:         imports,
		Exports:         exports,
		ServerID:        serverID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := uc.nodeRepo.Create(ctx, node); err != nil {
		return nil, err
	}

	return node, nil
}

func (uc *NodeUseCase) Get(ctx context.Context, id string) (*entity.Node, error) {
	return uc.nodeRepo.Get(ctx, id)
}

func (uc *NodeUseCase) List(ctx context.Context, offset, limit int) ([]*entity.Node, int64, error) {
	return uc.nodeRepo.List(ctx, offset, limit)
}

func (uc *NodeUseCase) Update(ctx context.Context, id, name, description string, nodeType entity.NodeType, priorityStartup int32, fields, imports, exports, serverID string) (*entity.Node, error) {
	node, err := uc.nodeRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	node.Name = name
	node.Description = description
	node.Type = nodeType
	node.PriorityStartup = priorityStartup
	node.Fields = fields
	node.Imports = imports
	node.Exports = exports
	node.ServerID = serverID
	node.UpdatedAt = time.Now()

	if err := uc.nodeRepo.Update(ctx, node); err != nil {
		return nil, err
	}

	return node, nil
}

func (uc *NodeUseCase) Delete(ctx context.Context, id string) error {
	// TODO: Add logic to stop container and cleanup data if running
	return uc.nodeRepo.Delete(ctx, id)
}

func (uc *NodeUseCase) Deploy(ctx context.Context, id string) error {
	// TODO: Implement deployment logic (call agent)
	return nil
}

func (uc *NodeUseCase) GetGraphStatus(ctx context.Context, id string) (string, string, error) {
	// TODO: Implement status check logic
	return "stopped", "", nil
}
