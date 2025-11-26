package grpc

import (
	"context"

	common "github.com/zhinea/sylix/internal/infra/proto/common"
	"github.com/zhinea/sylix/internal/infra/proto/controlplane"
	pbControlPlane "github.com/zhinea/sylix/internal/infra/proto/controlplane"
	services "github.com/zhinea/sylix/internal/module/controlplane/domain/services"
)

type NodeHandler struct {
	pbControlPlane.UnimplementedNodeServiceServer
	nodeService *services.NodeService
}

func NewNodeHandler(nodeService *services.NodeService) *NodeHandler {
	return &NodeHandler{
		nodeService: nodeService,
	}
}

func (h *NodeHandler) CreateNodeGraph(ctx context.Context, req *pbControlPlane.CreateNodeGraphRequest) (*pbControlPlane.NodeGraphResponse, error) {
	return h.nodeService.CreateNodeGraph(ctx, req)
}

func (h *NodeHandler) GetNodeGraph(ctx context.Context, req *controlplane.GetNodeGraphRequest) (*controlplane.NodeGraphResponse, error) {
	// TODO: Implement GetNodeGraph in service
	return &controlplane.NodeGraphResponse{
		Status: common.StatusCode_NOT_FOUND,
		Error:  "Not implemented",
	}, nil
}

func (h *NodeHandler) UpdateNodeGraph(ctx context.Context, req *controlplane.UpdateNodeGraphRequest) (*controlplane.NodeGraphResponse, error) {
	// TODO: Implement UpdateNodeGraph in service
	return &controlplane.NodeGraphResponse{
		Status: common.StatusCode_NOT_FOUND,
		Error:  "Not implemented",
	}, nil
}

func (h *NodeHandler) DeleteNodeGraph(ctx context.Context, req *controlplane.DeleteNodeGraphRequest) (*common.MessageResponse, error) {
	// TODO: Implement DeleteNodeGraph in service
	return &common.MessageResponse{
		Status:  common.StatusCode_NOT_FOUND,
		Message: "Not implemented",
	}, nil
}

func (h *NodeHandler) DeployNodeGraph(ctx context.Context, req *controlplane.DeployNodeGraphRequest) (*common.MessageResponse, error) {
	return h.nodeService.DeployNodeGraph(ctx, req)
}

func (h *NodeHandler) GetNodeGraphStatus(ctx context.Context, req *controlplane.GetNodeGraphStatusRequest) (*controlplane.GetNodeGraphStatusResponse, error) {
	// TODO: Implement GetNodeGraphStatus in service
	return &controlplane.GetNodeGraphStatusResponse{
		Status: "UNKNOWN",
	}, nil
}

func (h *NodeHandler) GetDeploymentLogs(ctx context.Context, req *controlplane.GetDeploymentLogsRequest) (*controlplane.GetDeploymentLogsResponse, error) {
	// TODO: Implement GetDeploymentLogs in service
	return &controlplane.GetDeploymentLogsResponse{
		Logs: []string{},
	}, nil
}
