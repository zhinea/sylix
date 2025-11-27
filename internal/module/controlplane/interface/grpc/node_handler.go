package grpc

import (
	"context"

	"github.com/zhinea/sylix/internal/infra/proto/common"
	pb "github.com/zhinea/sylix/internal/infra/proto/controlplane/nodes"
	"github.com/zhinea/sylix/internal/module/controlplane/app"
	"github.com/zhinea/sylix/internal/module/controlplane/entity"
)

type NodeService struct {
	pb.UnimplementedNodeServiceServer
	nodeUseCase *app.NodeUseCase
}

func NewNodeService(nodeUseCase *app.NodeUseCase) *NodeService {
	return &NodeService{
		nodeUseCase: nodeUseCase,
	}
}

func (s *NodeService) CreateNode(ctx context.Context, req *pb.CreateNodeRequest) (*pb.NodeResponse, error) {
	node, err := s.nodeUseCase.Create(ctx, req.Name, req.Description, entity.NodeType(req.Type), req.PriorityStartup, req.Fields, req.Imports, req.Exports, req.ServerId)
	if err != nil {
		return &pb.NodeResponse{
			Status: common.StatusCode_INTERNAL_ERROR,
			Error:  err.Error(),
		}, nil
	}

	return &pb.NodeResponse{
		Status: common.StatusCode_CREATED,
		Node:   toPbNode(node),
	}, nil
}

func (s *NodeService) GetNode(ctx context.Context, req *pb.GetNodeRequest) (*pb.NodeResponse, error) {
	node, err := s.nodeUseCase.Get(ctx, req.Id)
	if err != nil {
		return &pb.NodeResponse{
			Status: common.StatusCode_NOT_FOUND,
			Error:  err.Error(),
		}, nil
	}

	return &pb.NodeResponse{
		Status: common.StatusCode_OK,
		Node:   toPbNode(node),
	}, nil
}

func (s *NodeService) ListNodes(ctx context.Context, req *pb.ListNodesRequest) (*pb.ListNodesResponse, error) {
	offset := int((req.Page - 1) * req.PageSize)
	nodes, count, err := s.nodeUseCase.List(ctx, offset, int(req.PageSize))
	if err != nil {
		return &pb.ListNodesResponse{
			Status: common.StatusCode_INTERNAL_ERROR,
			Error:  err.Error(),
		}, nil
	}

	var pbNodes []*pb.Node
	for _, node := range nodes {
		pbNodes = append(pbNodes, toPbNode(node))
	}

	return &pb.ListNodesResponse{
		Status:     common.StatusCode_OK,
		Nodes:      pbNodes,
		TotalCount: count,
	}, nil
}

func (s *NodeService) UpdateNode(ctx context.Context, req *pb.UpdateNodeRequest) (*pb.NodeResponse, error) {
	node, err := s.nodeUseCase.Update(ctx, req.Id, req.Name, req.Description, entity.NodeType(req.Type), req.PriorityStartup, req.Fields, req.Imports, req.Exports, req.ServerId)
	if err != nil {
		return &pb.NodeResponse{
			Status: common.StatusCode_INTERNAL_ERROR,
			Error:  err.Error(),
		}, nil
	}

	return &pb.NodeResponse{
		Status: common.StatusCode_OK,
		Node:   toPbNode(node),
	}, nil
}

func (s *NodeService) DeleteNode(ctx context.Context, req *pb.DeleteNodeRequest) (*common.MessageResponse, error) {
	err := s.nodeUseCase.Delete(ctx, req.Id)
	if err != nil {
		return &common.MessageResponse{
			Status:  common.StatusCode_INTERNAL_ERROR,
			Message: err.Error(),
		}, nil
	}

	return &common.MessageResponse{
		Status:  common.StatusCode_OK,
		Message: "Node deleted successfully",
	}, nil
}

func (s *NodeService) DeployNode(ctx context.Context, req *pb.DeployNodeRequest) (*common.MessageResponse, error) {
	err := s.nodeUseCase.Deploy(ctx, req.Id)
	if err != nil {
		return &common.MessageResponse{
			Status:  common.StatusCode_INTERNAL_ERROR,
			Message: err.Error(),
		}, nil
	}

	return &common.MessageResponse{
		Status:  common.StatusCode_OK,
		Message: "Node deployment triggered",
	}, nil
}

func (s *NodeService) GetNodeGraphStatus(ctx context.Context, req *pb.GetNodeGraphStatusRequest) (*pb.GetNodeGraphStatusResponse, error) {
	status, logs, err := s.nodeUseCase.GetGraphStatus(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.GetNodeGraphStatusResponse{
		Status: status,
		Logs:   logs,
	}, nil
}

func toPbNode(node *entity.Node) *pb.Node {
	return &pb.Node{
		Id:              node.ID,
		Name:            node.Name,
		Description:     node.Description,
		Type:            string(node.Type),
		PriorityStartup: node.PriorityStartup,
		Fields:          node.Fields,
		Imports:         node.Imports,
		Exports:         node.Exports,
		ServerId:        node.ServerID,
		CreatedAt:       node.CreatedAt.String(),
		UpdatedAt:       node.UpdatedAt.String(),
	}
}
