package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
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
	neonService   *services.NeonService
}

func NewAgentUseCase(configPath string, dockerService *services.DockerService, neonService *services.NeonService) *AgentUseCase {
	return &AgentUseCase{
		startTime:     time.Now(),
		configPath:    configPath,
		dockerService: dockerService,
		neonService:   neonService,
	}
}

func (uc *AgentUseCase) CreateDatabase(ctx context.Context, req *pbAgent.CreateDatabaseRequest) (*pbAgent.CreateDatabaseResponse, error) {
	logger.Log.Info("Received CreateDatabase command",
		zap.String("name", req.Name),
		zap.String("db_name", req.DbName),
		zap.String("branch", req.Branch),
	)

	// Ensure Neon Infrastructure is running
	if err := uc.neonService.EnsureInfrastructure(ctx); err != nil {
		logger.Log.Error("Failed to ensure neon infrastructure", zap.Error(err))
		return nil, err
	}
	logger.Log.Info("Neon infrastructure is ready")

	// Create Tenant
	tenantID, err := uc.neonService.CreateTenant(ctx)
	if err != nil {
		logger.Log.Error("Failed to create tenant", zap.Error(err))
		return nil, err
	}
	logger.Log.Info("Created Neon Tenant", zap.String("tenant_id", tenantID))

	// Create Timeline
	branch := req.Branch
	if branch == "" {
		branch = "main"
	}
	timelineID, err := uc.neonService.CreateTimeline(ctx, tenantID, branch, "")
	if err != nil {
		logger.Log.Error("Failed to create timeline", zap.Error(err))
		return nil, err
	}
	logger.Log.Info("Created Neon Timeline", zap.String("timeline_id", timelineID))

	pgVersion := req.PgVersion
	if pgVersion == 0 {
		pgVersion = 16
	}
	logger.Log.Info("Using Postgres version", zap.Int32("version", pgVersion))
	imageName := fmt.Sprintf("ghcr.io/neondatabase/compute-node-v%d:latest", pgVersion)

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

	// Construct Postgres command with Neon extensions configuration
	cmd := []string{
		"postgres",
		"-c", "config_file=/var/db/postgres/postgresql.conf",
		"-c", "neon.pageserver_connstring=postgresql://no_user@pageserver:6400",
		"-c", fmt.Sprintf("neon.tenant_id=%s", tenantID),
		"-c", fmt.Sprintf("neon.timeline_id=%s", timelineID),
		"-c", "neon.safekeeper_connstring=postgresql://no_user@safekeeper1:5454",
		"-h", "0.0.0.0",
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

	networkingConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"sylix-neon": {},
		},
	}

	config := &container.Config{
		Image: imageName,
		Env:   env,
		Cmd:   cmd,
		ExposedPorts: nat.PortSet{
			nat.Port(port): struct{}{},
		},
	}

	containerName := "sylix-db-" + req.Name + "-" + req.Branch
	logger.Log.Debug("Creating container", zap.String("container_name", containerName))

	containerID, err := uc.dockerService.CreateAndStartContainer(ctx, config, hostConfig, networkingConfig, containerName)
	if err != nil {
		logger.Log.Error("Failed to create/start container", zap.Error(err), zap.String("container_name", containerName))
		return nil, err
	}

	return &pbAgent.CreateDatabaseResponse{
		ContainerId: containerID,
		Port:        5432, // TODO: Get actual mapped port
		Status:      "RUNNING",
		TenantId:    tenantID,
		TimelineId:  timelineID,
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
