package app

import (
	"context"
	"fmt"

	"github.com/zhinea/sylix/internal/common/logger"
	"github.com/zhinea/sylix/internal/module/controlplane/domain/repository"
	"github.com/zhinea/sylix/internal/module/controlplane/domain/services"
	"github.com/zhinea/sylix/internal/module/controlplane/entity"
	"go.uber.org/zap"
)

type ServerUseCase struct {
	repo              repository.ServerRepository
	monitoringService *services.MonitoringService
	nodeService       *services.NodeService
}

func NewServerUseCase(
	repo repository.ServerRepository,
	monitoringService *services.MonitoringService,
	nodeService *services.NodeService,
) *ServerUseCase {
	return &ServerUseCase{
		repo:              repo,
		monitoringService: monitoringService,
		nodeService:       nodeService,
	}
}

func (uc *ServerUseCase) Create(ctx context.Context, server *entity.Server) (*entity.Server, error) {
	// Set initial status
	server.Status = entity.ServerStatusDisconnected

	// Check connection before creating
	if err := uc.nodeService.CheckConnection(server); err == nil {
		server.Status = entity.ServerStatusConnected
	} else {
		server.Status = entity.ServerStatusDisconnected
		logger.Log.Warn("Failed to connect to server during creation", zap.Error(err), zap.String("ip", server.IpAddress))
	}

	server.Agent.Port = 8083

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
	server.Agent.Key = existing.Agent.Key

	// Preserve Password and SSHKey if not provided (nil)
	if server.Credential.Password == nil {
		server.Credential.Password = existing.Credential.Password
	}
	if server.Credential.SSHKey == nil {
		server.Credential.SSHKey = existing.Credential.SSHKey
	}

	// Check connection before updating
	if err := uc.nodeService.CheckConnection(server); err == nil {
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

	if err := uc.nodeService.CheckConnection(server); err == nil {
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

func (uc *ServerUseCase) ProvisionNode(ctx context.Context, server *entity.Server) error {
	if server.Status != entity.ServerStatusConnected {
		return fmt.Errorf("server must be connected to provision node")
	}

	// Start async provisioning
	go uc.nodeService.Install(context.Background(), server)

	return nil
}

func (uc *ServerUseCase) GetStats(ctx context.Context, serverID string) ([]*entity.ServerStat, error) {
	return uc.monitoringService.GetStats(ctx, serverID)
}

func (uc *ServerUseCase) GetRealtimeStats(ctx context.Context, serverID string, limit int) ([]*entity.ServerPing, error) {
	return uc.monitoringService.GetRealtimeStats(ctx, serverID, limit)
}
