package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/zhinea/sylix/internal/infra/proto/agent"
	"github.com/zhinea/sylix/internal/infra/proto/common"
	"github.com/zhinea/sylix/internal/infra/proto/controlplane"
	"github.com/zhinea/sylix/internal/module/controlplane/domain/repository"
	"github.com/zhinea/sylix/internal/module/controlplane/entity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/datatypes"
)

type NodeService struct {
	repo       repository.NodeRepository
	serverRepo repository.ServerRepository
}

func NewNodeService(repo repository.NodeRepository, serverRepo repository.ServerRepository) *NodeService {
	return &NodeService{
		repo:       repo,
		serverRepo: serverRepo,
	}
}

// Helper to get or create agent client
func (s *NodeService) getAgentClient(ctx context.Context, serverID string) (agent.AgentClient, error) {
	server, err := s.serverRepo.GetByID(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to get server: %w", err)
	}

	// Assuming Agent Port is stored in server.Agent.Port
	// If it's 0, default to 9090 or similar if not specified
	port := server.Agent.Port
	if port == 0 {
		port = 9090 // Default agent port
	}

	target := fmt.Sprintf("%s:%d", server.IpAddress, port)

	// In a real app, we should cache these connections
	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial agent: %w", err)
	}

	return agent.NewAgentClient(conn), nil
}

func (s *NodeService) CreateNodeGraph(ctx context.Context, req *controlplane.CreateNodeGraphRequest) (*controlplane.NodeGraphResponse, error) {
	// Convert Proto nodes/edges to Entity JSON
	nodesJSON, _ := json.Marshal(req.Nodes)
	edgesJSON, _ := json.Marshal(req.Edges)

	graph := &entity.NodeGraph{
		ID:    fmt.Sprintf("graph-%d", time.Now().UnixNano()), // Simple ID generation
		Name:  req.Name,
		Nodes: datatypes.JSON(nodesJSON),
		Edges: datatypes.JSON(edgesJSON),
	}

	if err := s.repo.Create(ctx, graph); err != nil {
		return &controlplane.NodeGraphResponse{
			Status: common.StatusCode_INTERNAL_ERROR,
			Error:  err.Error(),
		}, nil
	}

	return &controlplane.NodeGraphResponse{
		Status: common.StatusCode_CREATED,
		Graph:  mapEntityToProto(graph),
	}, nil
}

func (s *NodeService) DeployNodeGraph(ctx context.Context, req *controlplane.DeployNodeGraphRequest) (*common.MessageResponse, error) {
	graph, err := s.repo.Get(ctx, req.Id)
	if err != nil {
		return &common.MessageResponse{
			Status:  common.StatusCode_NOT_FOUND,
			Message: "Graph not found",
		}, nil
	}

	// 1. Parse Graph
	var nodes []controlplane.Node
	var edges []controlplane.Edge
	json.Unmarshal(graph.Nodes, &nodes)
	json.Unmarshal(graph.Edges, &edges)

	// 2. Group by ServerID
	nodesByServer := make(map[string][]controlplane.Node)
	for _, node := range nodes {
		serverID := node.Data.ServerId
		nodesByServer[serverID] = append(nodesByServer[serverID], node)
	}

	// 3. Generate Compose and Deploy per Server
	for serverID, serverNodes := range nodesByServer {
		composeContent, err := generateDockerCompose(serverNodes, edges)
		if err != nil {
			return &common.MessageResponse{
				Status:  common.StatusCode_INTERNAL_ERROR,
				Message: fmt.Sprintf("Failed to generate compose for server %s: %v", serverID, err),
			}, nil
		}

		client, err := s.getAgentClient(ctx, serverID)
		if err != nil {
			// Log error but maybe continue or fail? For now fail.
			return &common.MessageResponse{
				Status:  common.StatusCode_INTERNAL_ERROR,
				Message: fmt.Sprintf("Failed to connect to agent on server %s: %v", serverID, err),
			}, nil
		}

		// Project name based on graph ID
		projectName := fmt.Sprintf("sylix-%s", graph.ID)

		_, err = client.DeployCompose(ctx, &agent.DeployComposeRequest{
			ProjectName:    projectName,
			ComposeContent: composeContent,
		})
		if err != nil {
			return &common.MessageResponse{
				Status:  common.StatusCode_INTERNAL_ERROR,
				Message: fmt.Sprintf("Agent deployment failed on server %s: %v", serverID, err),
			}, nil
		}
	}

	return &common.MessageResponse{
		Status:  common.StatusCode_OK,
		Message: "Deployment initiated successfully",
	}, nil
}

// Helper functions
func mapEntityToProto(e *entity.NodeGraph) *controlplane.NodeGraph {
	var nodes []*controlplane.Node
	var edges []*controlplane.Edge
	json.Unmarshal(e.Nodes, &nodes)
	json.Unmarshal(e.Edges, &edges)

	return &controlplane.NodeGraph{
		Id:        e.ID,
		Name:      e.Name,
		Nodes:     nodes,
		Edges:     edges,
		CreatedAt: e.CreatedAt.String(),
		UpdatedAt: e.UpdatedAt.String(),
	}
}

func generateDockerCompose(nodes []controlplane.Node, edges []controlplane.Edge) (string, error) {
	// Sort nodes by priority (Storage Broker > Safekeeper > Pageserver > Compute)
	sort.Slice(nodes, func(i, j int) bool {
		return getPriority(nodes[i].Type) < getPriority(nodes[j].Type)
	})

	var services strings.Builder
	services.WriteString("version: '3.8'\nservices:\n")

	for _, node := range nodes {
		image := getImageForType(node.Type)

		// Basic service definition
		services.WriteString(fmt.Sprintf("  %s:\n", node.Label))
		services.WriteString(fmt.Sprintf("    image: %s\n", image))
		services.WriteString("    restart: always\n")

		// Environment variables
		services.WriteString("    environment:\n")
		services.WriteString(fmt.Sprintf("      - NODE_ID=%s\n", node.Id))
		if node.Data.PgVersion != "" {
			services.WriteString(fmt.Sprintf("      - PG_VERSION=%s\n", node.Data.PgVersion))
		}

		// Add connections based on edges
		// Find edges where this node is the source (outgoing connection)
		// e.g. Compute -> Pageserver
		for _, edge := range edges {
			if edge.Source == node.Id {
				// Find target node to get its label/hostname
				var targetNode *controlplane.Node
				for _, n := range nodes {
					if n.Id == edge.Target {
						targetNode = &n
						break
					}
				}
				if targetNode != nil {
					// Env var name convention: TARGET_TYPE_HOST
					envName := fmt.Sprintf("%s_HOST", strings.ToUpper(targetNode.Type))
					services.WriteString(fmt.Sprintf("      - %s=%s\n", envName, targetNode.Label))
				}
			}
		}

		// Ports
		if node.Data.PgPort > 0 {
			services.WriteString("    ports:\n")
			services.WriteString(fmt.Sprintf("      - \"%d:5432\"\n", node.Data.PgPort))
		}
	}

	return services.String(), nil
}

func getImageForType(nodeType string) string {
	switch nodeType {
	case "compute":
		return "neondatabase/compute-node:latest"
	case "pageserver":
		return "neondatabase/pageserver:latest"
	case "safekeeper":
		return "neondatabase/safekeeper:latest"
	case "storage_broker":
		return "neondatabase/storage-broker:latest"
	default:
		return "postgres:15-alpine"
	}
}

func getPriority(nodeType string) int {
	switch nodeType {
	case "storage_broker":
		return 1
	case "safekeeper":
		return 2
	case "pageserver":
		return 3
	case "compute":
		return 4
	default:
		return 99
	}
}
