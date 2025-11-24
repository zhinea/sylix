package grpc

import (
	"context"

	"github.com/zhinea/sylix/internal/common/logger"
	pbCommon "github.com/zhinea/sylix/internal/infra/proto/common"
	pb "github.com/zhinea/sylix/internal/infra/proto/controlplane"
	"github.com/zhinea/sylix/internal/module/controlplane/domain/services"
	"github.com/zhinea/sylix/internal/module/controlplane/entity"
	"go.uber.org/zap"
)

type DatabaseHandler struct {
	pb.UnimplementedDatabaseServiceServer
	service *services.DatabaseService
}

func NewDatabaseHandler(service *services.DatabaseService) *DatabaseHandler {
	return &DatabaseHandler{
		service: service,
	}
}

func (h *DatabaseHandler) Create(ctx context.Context, req *pb.Database) (*pb.DatabaseResponse, error) {
	logger.Log.Info("Received CreateDatabase request",
		zap.String("name", req.Name),
		zap.String("server_id", req.ServerId),
		zap.String("db_name", req.DbName),
		zap.String("branch", req.Branch),
	)

	db := &entity.Database{
		Name:     req.Name,
		User:     req.User,
		Password: req.Password,
		DbName:   req.DbName,
		Branch:   req.Branch,
		ServerID: req.ServerId,
	}

	createdDb, err := h.service.Create(ctx, db)
	if err != nil {
		logger.Log.Error("Failed to create database via handler", zap.Error(err))
		return nil, err
	}

	return &pb.DatabaseResponse{
		Database: h.toProto(createdDb),
	}, nil
}

func (h *DatabaseHandler) Get(ctx context.Context, req *pb.DatabaseId) (*pb.DatabaseResponse, error) {
	db, err := h.service.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.DatabaseResponse{
		Database: h.toProto(db),
	}, nil
}

func (h *DatabaseHandler) All(ctx context.Context, req *pbCommon.Empty) (*pb.DatabasesResponse, error) {
	dbs, err := h.service.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var protoDbs []*pb.Database
	for _, db := range dbs {
		protoDbs = append(protoDbs, h.toProto(db))
	}

	return &pb.DatabasesResponse{
		Databases: protoDbs,
	}, nil
}

func (h *DatabaseHandler) Delete(ctx context.Context, req *pb.DatabaseId) (*pb.DatabaseMessageResponse, error) {
	err := h.service.Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.DatabaseMessageResponse{
		Status:  200,
		Message: "Deleted successfully",
	}, nil
}

func (h *DatabaseHandler) toProto(db *entity.Database) *pb.Database {
	return &pb.Database{
		Id:          db.Id,
		Name:        db.Name,
		User:        db.User,
		Password:    db.Password,
		DbName:      db.DbName,
		Branch:      db.Branch,
		ServerId:    db.ServerID,
		Status:      db.Status,
		ContainerId: db.ContainerID,
		Port:        int32(db.Port),
	}
}
