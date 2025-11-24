package services

import (
	"context"
	"errors"

	"github.com/zhinea/sylix/internal/common/logger"
	pbAgent "github.com/zhinea/sylix/internal/infra/proto/agent"
	"github.com/zhinea/sylix/internal/module/controlplane/domain/repository"
	"github.com/zhinea/sylix/internal/module/controlplane/entity"
	"go.uber.org/zap"
)

type DatabaseService struct {
	repo         repository.DatabaseRepository
	serverRepo   repository.ServerRepository
	agentService *AgentService
}

func NewDatabaseService(repo repository.DatabaseRepository, serverRepo repository.ServerRepository, agentService *AgentService) *DatabaseService {
	return &DatabaseService{
		repo:         repo,
		serverRepo:   serverRepo,
		agentService: agentService,
	}
}

func (s *DatabaseService) Create(ctx context.Context, db *entity.Database) (*entity.Database, error) {
	logger.Log.Info("Starting database creation flow",
		zap.String("name", db.Name),
		zap.String("server_id", db.ServerID),
		zap.String("branch", db.Branch),
	)

	// 1. Validate Server
	server, err := s.serverRepo.GetByID(ctx, db.ServerID)
	if err != nil {
		logger.Log.Error("Failed to fetch server", zap.Error(err), zap.String("server_id", db.ServerID))
		return nil, err
	}
	if server == nil {
		logger.Log.Error("Server not found", zap.String("server_id", db.ServerID))
		return nil, errors.New("server not found")
	}

	// 2. Save initial state
	db.Status = entity.DatabaseStatusCreating
	if db.Branch == "" {
		db.Branch = "main"
	}
	createdDb, err := s.repo.Create(ctx, db)
	if err != nil {
		logger.Log.Error("Failed to create initial database record", zap.Error(err))
		return nil, err
	}

	logger.Log.Debug("Initial database record created", zap.String("db_id", createdDb.Id))

	// 3. Call Agent
	req := &pbAgent.CreateDatabaseRequest{
		Name:     db.Name,
		User:     db.User,
		Password: db.Password,
		DbName:   db.DbName,
		Branch:   db.Branch,
	}

	logger.Log.Info("Delegating to agent for container creation", zap.String("server_ip", server.IpAddress))
	resp, err := s.agentService.CreateDatabase(ctx, server, req)
	if err != nil {
		logger.Log.Error("Agent failed to create database container",
			zap.Error(err),
			zap.String("server_id", server.Id),
		)
		createdDb.Status = entity.DatabaseStatusError
		s.repo.Update(ctx, createdDb)
		return nil, err
	}

	// 4. Update with result
	createdDb.Status = entity.DatabaseStatusRunning
	createdDb.ContainerID = resp.ContainerId
	createdDb.Port = int(resp.Port)

	updatedDb, err := s.repo.Update(ctx, createdDb)
	if err != nil {
		logger.Log.Error("Failed to update database record with container info", zap.Error(err))
		return nil, err
	}

	logger.Log.Info("Database creation completed successfully",
		zap.String("db_id", updatedDb.Id),
		zap.String("container_id", updatedDb.ContainerID),
		zap.Int("port", updatedDb.Port),
	)

	return updatedDb, nil
}

func (s *DatabaseService) GetByID(ctx context.Context, id string) (*entity.Database, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *DatabaseService) GetAll(ctx context.Context) ([]*entity.Database, error) {
	return s.repo.GetAll(ctx)
}

func (s *DatabaseService) Delete(ctx context.Context, id string) error {
	// TODO: Call agent to stop/remove container
	return s.repo.Delete(ctx, id)
}
