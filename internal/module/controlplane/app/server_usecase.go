package app

import (
	"context"
	"fmt"
	"time"

	"github.com/zhinea/sylix/internal/common/logger"
	"github.com/zhinea/sylix/internal/module/controlplane/domain/repository"
	"github.com/zhinea/sylix/internal/module/controlplane/domain/services"
	"github.com/zhinea/sylix/internal/module/controlplane/entity"
	"go.uber.org/zap"
)

type ServerUseCase struct {
	repo              repository.ServerRepository
	monitoringService *services.MonitoringService
	agentService      *services.AgentService
}

func NewServerUseCase(
	repo repository.ServerRepository,
	monitoringService *services.MonitoringService,
	agentService *services.AgentService,
) *ServerUseCase {
	return &ServerUseCase{
		repo:              repo,
		monitoringService: monitoringService,
		agentService:      agentService,
	}
}

func (uc *ServerUseCase) Create(ctx context.Context, server *entity.Server) (*entity.Server, error) {
	// Set initial status
	server.Status = entity.ServerStatusDisconnected

	// Check connection before creating
	if err := uc.agentService.CheckConnection(server); err == nil {
		server.Status = entity.ServerStatusConnected
	} else {
		server.Status = entity.ServerStatusDisconnected
		logger.Log.Warn("Failed to connect to server during creation", zap.Error(err), zap.String("ip", server.IpAddress))
	}

	return uc.repo.Create(ctx, server)
}

func (uc *ServerUseCase) Get(ctx context.Context, id string) (*entity.Server, error) {
	server, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return server, nil
}

func (uc *ServerUseCase) GetAll(ctx context.Context) ([]*entity.Server, error) {
	return uc.repo.GetAll(ctx)
}

func (uc *ServerUseCase) Update(ctx context.Context, server *entity.Server) (*entity.Server, error) {
	// Fetch existing server to preserve internal fields
	existing, err := uc.repo.GetByID(ctx, server.Id)
	if err != nil {
		return nil, err
	}

	// Preserve Cert
	server.Agent.Cert = existing.Agent.Cert

	// Preserve Password and SSHKey if not provided (nil)
	if server.Credential.Password == nil {
		server.Credential.Password = existing.Credential.Password
	}
	if server.Credential.SSHKey == nil {
		server.Credential.SSHKey = existing.Credential.SSHKey
	}

	// Check connection before updating
	if err := uc.agentService.CheckConnection(server); err == nil {
		server.Status = entity.ServerStatusConnected
	} else {
		server.Status = entity.ServerStatusDisconnected
		logger.Log.Warn("Failed to connect to server during update", zap.Error(err), zap.String("ip", server.IpAddress))
	}

	return uc.repo.Update(ctx, server)
}

func (uc *ServerUseCase) RetryConnection(ctx context.Context, id string) (*entity.Server, error) {
	server, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := uc.agentService.CheckConnection(server); err == nil {
		server.Status = entity.ServerStatusConnected
	} else {
		server.Status = entity.ServerStatusDisconnected
		logger.Log.Warn("Failed to connect to server during retry", zap.Error(err), zap.String("ip", server.IpAddress))
	}

	return uc.repo.Update(ctx, server)
}

func (uc *ServerUseCase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *ServerUseCase) GetAgentConfig(ctx context.Context, id string) (string, string, error) {
	server, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return "", "", err
	}
	return uc.agentService.GetAgentConfig(ctx, server)
}

func (uc *ServerUseCase) InstallAgent(ctx context.Context, serverID string) error {
	server, err := uc.repo.GetByID(ctx, serverID)
	if err != nil {
		return fmt.Errorf("failed to get server: %w", err)
	}

	if server.Status != entity.ServerStatusConnected {
		return fmt.Errorf("server must be connected to install agent")
	}

	// Start async installation
	go uc.agentService.Install(context.Background(), server)

	return nil
}

func (uc *ServerUseCase) GetStats(ctx context.Context, serverID string) ([]*entity.ServerStat, error) {
	return uc.monitoringService.GetStats(ctx, serverID)
}

func (uc *ServerUseCase) GetRealtimeStats(ctx context.Context, serverID string, limit int) ([]*entity.ServerPing, error) {
	return uc.monitoringService.GetRealtimeStats(ctx, serverID, limit)
}

func (uc *ServerUseCase) GetAccidents(ctx context.Context, serverID string, startDate, endDate *time.Time, resolved *bool, page, pageSize int) ([]*entity.ServerAccident, int64, error) {
	return uc.monitoringService.GetAccidents(ctx, serverID, startDate, endDate, resolved, page, pageSize)
}

func (uc *ServerUseCase) ConfigureAgent(ctx context.Context, serverID string, configStr string) error {
	server, err := uc.repo.GetByID(ctx, serverID)
	if err != nil {
		return err
	}
	return uc.agentService.Configure(ctx, server, configStr)
}

func (uc *ServerUseCase) UpdateAgentPort(ctx context.Context, serverID string, port int) error {
	server, err := uc.repo.GetByID(ctx, serverID)
	if err != nil {
		return err
	}
	return uc.agentService.UpdatePort(ctx, server, port)
}

func (uc *ServerUseCase) UpdateServerTimeZone(ctx context.Context, serverID string, timezone string) error {
	server, err := uc.repo.GetByID(ctx, serverID)
	if err != nil {
		return err
	}
	return uc.agentService.UpdateTimeZone(ctx, server, timezone)
}

func (uc *ServerUseCase) DeleteAccident(ctx context.Context, id string) error {
	return uc.monitoringService.DeleteAccident(ctx, id)
}

func (uc *ServerUseCase) BatchDeleteAccidents(ctx context.Context, ids []string) error {
	return uc.monitoringService.BatchDeleteAccidents(ctx, ids)
}
