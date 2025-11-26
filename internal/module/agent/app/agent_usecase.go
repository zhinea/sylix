package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	pbAgent "github.com/zhinea/sylix/internal/infra/proto/agent"
	"github.com/zhinea/sylix/internal/infra/proto/common"
)

type AgentUseCase struct {
	startTime  time.Time
	configPath string
	baseDir    string
}

func NewAgentUseCase(configPath string) *AgentUseCase {
	return &AgentUseCase{
		startTime:  time.Now(),
		configPath: configPath,
		baseDir:    "/var/lib/sylix/compose", // Default base dir
	}
}

func (uc *AgentUseCase) GetStatus(ctx context.Context) (*pbAgent.GetStatusResponse, error) {
	return &pbAgent.GetStatusResponse{
		Status:  "RUNNING",
		Version: "0.1.0",
		Uptime:  int64(time.Since(uc.startTime).Seconds()),
	}, nil
}

func (uc *AgentUseCase) Heartbeat(ctx context.Context) (*pbAgent.HeartbeatResponse, error) {
	return &pbAgent.HeartbeatResponse{
		Acknowledged: true,
	}, nil
}

func (uc *AgentUseCase) Ping(ctx context.Context) (*pbAgent.PingResponse, error) {
	return &pbAgent.PingResponse{
		Timestamp: time.Now().Unix(),
		Status:    "OK",
	}, nil
}

func (uc *AgentUseCase) GetConfig(ctx context.Context) (*pbAgent.GetConfigResponse, error) {
	configContent, err := os.ReadFile(uc.configPath)
	if err != nil {
		return nil, err
	}
	return &pbAgent.GetConfigResponse{
		Config: string(configContent),
	}, nil
}

func (uc *AgentUseCase) DeployCompose(ctx context.Context, req *pbAgent.DeployComposeRequest) (*common.MessageResponse, error) {
	projectDir := filepath.Join(uc.baseDir, req.ProjectName)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return &common.MessageResponse{
			Status:  common.StatusCode_INTERNAL_ERROR,
			Message: fmt.Sprintf("Failed to create directory: %v", err),
		}, nil
	}

	composeFile := filepath.Join(projectDir, "docker-compose.yml")
	if err := os.WriteFile(composeFile, []byte(req.ComposeContent), 0644); err != nil {
		return &common.MessageResponse{
			Status:  common.StatusCode_INTERNAL_ERROR,
			Message: fmt.Sprintf("Failed to write compose file: %v", err),
		}, nil
	}

	cmd := exec.CommandContext(ctx, "docker-compose", "-f", composeFile, "-p", req.ProjectName, "up", "-d")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &common.MessageResponse{
			Status:  common.StatusCode_INTERNAL_ERROR,
			Message: fmt.Sprintf("Failed to deploy: %v, output: %s", err, string(output)),
		}, nil
	}

	return &common.MessageResponse{
		Status:  common.StatusCode_OK,
		Message: "Deployed successfully",
	}, nil
}

func (uc *AgentUseCase) StopCompose(ctx context.Context, req *pbAgent.StopComposeRequest) (*common.MessageResponse, error) {
	projectDir := filepath.Join(uc.baseDir, req.ProjectName)
	composeFile := filepath.Join(projectDir, "docker-compose.yml")

	cmd := exec.CommandContext(ctx, "docker-compose", "-f", composeFile, "-p", req.ProjectName, "down")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &common.MessageResponse{
			Status:  common.StatusCode_INTERNAL_ERROR,
			Message: fmt.Sprintf("Failed to stop: %v, output: %s", err, string(output)),
		}, nil
	}

	return &common.MessageResponse{
		Status:  common.StatusCode_OK,
		Message: "Stopped successfully",
	}, nil
}

func (uc *AgentUseCase) GetComposeStatus(ctx context.Context, req *pbAgent.GetComposeStatusRequest) (*pbAgent.GetComposeStatusResponse, error) {
	projectDir := filepath.Join(uc.baseDir, req.ProjectName)
	composeFile := filepath.Join(projectDir, "docker-compose.yml")

	cmd := exec.CommandContext(ctx, "docker-compose", "-f", composeFile, "-p", req.ProjectName, "ps", "--services", "--filter", "status=running")
	output, err := cmd.Output()
	if err != nil {
		// If command fails, maybe project doesn't exist or no running services
		return &pbAgent.GetComposeStatusResponse{
			Running:  false,
			Services: []string{},
		}, nil
	}

	services := strings.Split(strings.TrimSpace(string(output)), "\n")
	running := len(services) > 0 && services[0] != ""

	return &pbAgent.GetComposeStatusResponse{
		Running:  running,
		Services: services,
	}, nil
}

func (uc *AgentUseCase) GetComposeLogs(req *pbAgent.GetComposeLogsRequest, stream pbAgent.Agent_GetComposeLogsServer) error {
	projectDir := filepath.Join(uc.baseDir, req.ProjectName)
	composeFile := filepath.Join(projectDir, "docker-compose.yml")

	args := []string{"-f", composeFile, "-p", req.ProjectName, "logs"}
	if req.Follow {
		args = append(args, "-f")
	}
	if req.Tail > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", req.Tail))
	}

	cmd := exec.Command("docker-compose", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	// Merge stderr into stdout
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		if err := stream.Send(&pbAgent.GetComposeLogsResponse{LogLine: scanner.Text()}); err != nil {
			return err
		}
	}

	return cmd.Wait()
}
