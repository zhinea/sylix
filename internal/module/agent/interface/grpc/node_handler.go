package grpc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/zhinea/sylix/internal/infra/proto/agent"
	"github.com/zhinea/sylix/internal/infra/proto/common"
)

type AgentNodeService struct {
	agent.UnimplementedAgentNodeServiceServer
	composeDir string
}

func NewAgentNodeService(composeDir string) *AgentNodeService {
	return &AgentNodeService{
		composeDir: composeDir,
	}
}

func (s *AgentNodeService) DeployNodeCompose(req *agent.DeployNodeComposeRequest, stream agent.AgentNodeService_DeployNodeComposeServer) error {
	// Create compose directory if it doesn't exist
	nodeDir := filepath.Join(s.composeDir, req.NodeId)
	if err := os.MkdirAll(nodeDir, 0755); err != nil {
		stream.Send(&agent.DeployNodeComposeResponse{
			Log:    fmt.Sprintf("Error creating directory: %v", err),
			Status: "failed",
		})
		return err
	}

	// Write compose file
	composeFile := filepath.Join(nodeDir, "docker-compose.yml")
	if err := os.WriteFile(composeFile, []byte(req.ComposeContent), 0644); err != nil {
		stream.Send(&agent.DeployNodeComposeResponse{
			Log:    fmt.Sprintf("Error writing compose file: %v", err),
			Status: "failed",
		})
		return err
	}

	stream.Send(&agent.DeployNodeComposeResponse{
		Log:    fmt.Sprintf("Created compose file at %s", composeFile),
		Status: "running",
	})

	// Execute docker-compose up
	cmd := exec.CommandContext(stream.Context(), "docker-compose", "-f", composeFile, "up", "-d")
	cmd.Dir = nodeDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		stream.Send(&agent.DeployNodeComposeResponse{
			Log:    fmt.Sprintf("Error deploying: %v\nOutput: %s", err, string(output)),
			Status: "failed",
		})
		return err
	}

	stream.Send(&agent.DeployNodeComposeResponse{
		Log:    string(output),
		Status: "running",
	})

	// Final success message
	stream.Send(&agent.DeployNodeComposeResponse{
		Log:    "Deployment completed successfully",
		Status: "success",
	})

	return nil
}

func (s *AgentNodeService) StopNodeCompose(ctx context.Context, req *agent.StopNodeComposeRequest) (*common.MessageResponse, error) {
	nodeDir := filepath.Join(s.composeDir, req.NodeId)
	composeFile := filepath.Join(nodeDir, "docker-compose.yml")

	if _, err := os.Stat(composeFile); os.IsNotExist(err) {
		return &common.MessageResponse{
			Status:  common.StatusCode_NOT_FOUND,
			Message: "Node compose file not found",
		}, nil
	}

	cmd := exec.CommandContext(ctx, "docker-compose", "-f", composeFile, "down", "-v")
	cmd.Dir = nodeDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return &common.MessageResponse{
			Status:  common.StatusCode_INTERNAL_ERROR,
			Message: fmt.Sprintf("Error stopping node: %v\nOutput: %s", err, string(output)),
		}, nil
	}

	// Optionally remove the directory
	os.RemoveAll(nodeDir)

	return &common.MessageResponse{
		Status:  common.StatusCode_OK,
		Message: "Node stopped successfully",
	}, nil
}

func (s *AgentNodeService) GetNodeStatus(ctx context.Context, req *agent.GetNodeStatusRequest) (*agent.GetNodeStatusResponse, error) {
	nodeDir := filepath.Join(s.composeDir, req.NodeId)
	composeFile := filepath.Join(nodeDir, "docker-compose.yml")

	if _, err := os.Stat(composeFile); os.IsNotExist(err) {
		return &agent.GetNodeStatusResponse{
			Status: "unknown",
			Error:  "Node compose file not found",
		}, nil
	}

	cmd := exec.CommandContext(ctx, "docker-compose", "-f", composeFile, "ps", "-q")
	cmd.Dir = nodeDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return &agent.GetNodeStatusResponse{
			Status: "failed",
			Error:  err.Error(),
		}, nil
	}

	// If no containers are running, status is stopped
	if strings.TrimSpace(string(output)) == "" {
		return &agent.GetNodeStatusResponse{
			Status: "stopped",
		}, nil
	}

	return &agent.GetNodeStatusResponse{
		Status: "running",
	}, nil
}

func (s *AgentNodeService) GetNodeLogs(req *agent.GetNodeLogsRequest, stream agent.AgentNodeService_GetNodeLogsServer) error {
	nodeDir := filepath.Join(s.composeDir, req.NodeId)
	composeFile := filepath.Join(nodeDir, "docker-compose.yml")

	if _, err := os.Stat(composeFile); os.IsNotExist(err) {
		return fmt.Errorf("node compose file not found")
	}

	args := []string{"-f", composeFile, "logs"}
	if req.Tail > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", req.Tail))
	}

	cmd := exec.CommandContext(stream.Context(), "docker-compose", args...)
	cmd.Dir = nodeDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	// Send logs line by line
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line != "" {
			if err := stream.Send(&agent.NodeLogResponse{
				Log: line,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}
