package app

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/zhinea/sylix/internal/common/logger"
	pbAgent "github.com/zhinea/sylix/internal/infra/proto/agent"
	"github.com/zhinea/sylix/internal/module/agent/domain/services"
	"go.uber.org/zap"
)

type AgentUseCase struct {
	startTime     time.Time
	configPath    string
	dockerService *services.DockerService
}

func NewAgentUseCase(configPath string, dockerService *services.DockerService) *AgentUseCase {
	return &AgentUseCase{
		startTime:     time.Now(),
		configPath:    configPath,
		dockerService: dockerService,
	}
}

func (uc *AgentUseCase) CreateDatabase(ctx context.Context, req *pbAgent.CreateDatabaseRequest) (*pbAgent.CreateDatabaseResponse, error) {
	logger.Log.Info("Received CreateDatabase command",
		zap.String("name", req.Name),
		zap.String("db_name", req.DbName),
		zap.String("branch", req.Branch),
	)

	imageName := "postgres:16"
	logger.Log.Debug("Pulling image", zap.String("image", imageName))
	if err := uc.dockerService.PullImage(ctx, imageName); err != nil {
		logger.Log.Error("Failed to pull image", zap.Error(err), zap.String("image", imageName))
		return nil, err
	}

	env := []string{
		"POSTGRES_USER=" + req.User,
		"POSTGRES_PASSWORD=" + req.Password,
		"POSTGRES_DB=" + req.DbName,
	}

	port := "5432/tcp"
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(port): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "0", // Random available port
				},
			},
		},
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	config := &container.Config{
		Image: imageName,
		Env:   env,
		ExposedPorts: nat.PortSet{
			nat.Port(port): struct{}{},
		},
	}

	containerName := "sylix-db-" + req.Name + "-" + req.Branch
	logger.Log.Debug("Creating container", zap.String("container_name", containerName))

	containerID, err := uc.dockerService.CreateAndStartContainer(ctx, config, hostConfig, containerName)
	if err != nil {
		logger.Log.Error("Failed to create/start container", zap.Error(err), zap.String("container_name", containerName))
		return nil, err
	}

	hostPort, err := uc.dockerService.GetContainerPort(ctx, containerID, port)
	if err != nil {
		logger.Log.Error("Failed to get container port", zap.Error(err), zap.String("container_id", containerID))
		return nil, err
	}

	p, err := strconv.Atoi(hostPort)
	if err != nil {
		logger.Log.Error("Failed to parse port", zap.Error(err), zap.String("port_str", hostPort))
		return nil, err
	}

	logger.Log.Info("Database container started successfully",
		zap.String("container_id", containerID),
		zap.Int("port", p),
	)

	return &pbAgent.CreateDatabaseResponse{
		ContainerId: containerID,
		Port:        int32(p),
		Status:      "RUNNING",
	}, nil
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
