package grpc

import (
	"context"

	"github.com/zhinea/sylix/internal/common/model"
	pbCommon "github.com/zhinea/sylix/internal/infra/proto/common"
	pbControlPlane "github.com/zhinea/sylix/internal/infra/proto/controlplane"
	"github.com/zhinea/sylix/internal/module/controlplane/domain/services"
	"github.com/zhinea/sylix/internal/module/controlplane/entity"
)

type BackupStorageService struct {
	pbControlPlane.UnimplementedBackupStorageServiceServer
	service *services.BackupService
}

func NewBackupStorageService(service *services.BackupService) *BackupStorageService {
	return &BackupStorageService{
		service: service,
	}
}

func (s *BackupStorageService) Create(ctx context.Context, req *pbControlPlane.BackupStorage) (*pbControlPlane.BackupStorageResponse, error) {
	backup := s.protoToEntity(req)
	createdBackup, err := s.service.Create(ctx, backup)
	if err != nil {
		errStr := err.Error()
		return &pbControlPlane.BackupStorageResponse{
			Status: pbControlPlane.BackupStatusCode_BACKUP_INTERNAL_ERROR,
			Error:  &errStr,
		}, nil
	}
	return &pbControlPlane.BackupStorageResponse{
		Status: pbControlPlane.BackupStatusCode_BACKUP_CREATED,
		Data:   s.entityToProto(createdBackup),
	}, nil
}

func (s *BackupStorageService) Get(ctx context.Context, req *pbControlPlane.BackupStorageId) (*pbControlPlane.BackupStorageResponse, error) {
	backup, err := s.service.GetByID(ctx, req.Id)
	if err != nil {
		errStr := err.Error()
		return &pbControlPlane.BackupStorageResponse{
			Status: pbControlPlane.BackupStatusCode_BACKUP_NOT_FOUND,
			Error:  &errStr,
		}, nil
	}
	return &pbControlPlane.BackupStorageResponse{
		Status: pbControlPlane.BackupStatusCode_BACKUP_OK,
		Data:   s.entityToProto(backup),
	}, nil
}

func (s *BackupStorageService) All(ctx context.Context, _ *pbCommon.Empty) (*pbControlPlane.BackupStoragesResponse, error) {
	backups, err := s.service.GetAll(ctx)
	if err != nil {
		return &pbControlPlane.BackupStoragesResponse{
			Status: pbControlPlane.BackupStatusCode_BACKUP_INTERNAL_ERROR,
		}, nil
	}

	var pbBackups []*pbControlPlane.BackupStorage
	for _, b := range backups {
		pbBackups = append(pbBackups, s.entityToProto(b))
	}

	return &pbControlPlane.BackupStoragesResponse{
		Status: pbControlPlane.BackupStatusCode_BACKUP_OK,
		Data:   pbBackups,
	}, nil
}

func (s *BackupStorageService) Update(ctx context.Context, req *pbControlPlane.BackupStorage) (*pbControlPlane.BackupStorageResponse, error) {
	backup := s.protoToEntity(req)
	updatedBackup, err := s.service.Update(ctx, backup)
	if err != nil {
		errStr := err.Error()
		return &pbControlPlane.BackupStorageResponse{
			Status: pbControlPlane.BackupStatusCode_BACKUP_INTERNAL_ERROR,
			Error:  &errStr,
		}, nil
	}
	return &pbControlPlane.BackupStorageResponse{
		Status: pbControlPlane.BackupStatusCode_BACKUP_OK,
		Data:   s.entityToProto(updatedBackup),
	}, nil
}

func (s *BackupStorageService) Delete(ctx context.Context, req *pbControlPlane.BackupStorageId) (*pbControlPlane.BackupMessageResponse, error) {
	err := s.service.Delete(ctx, req.Id)
	if err != nil {
		return &pbControlPlane.BackupMessageResponse{
			Status:  pbControlPlane.BackupStatusCode_BACKUP_INTERNAL_ERROR,
			Message: err.Error(),
		}, nil
	}
	return &pbControlPlane.BackupMessageResponse{
		Status:  pbControlPlane.BackupStatusCode_BACKUP_OK,
		Message: "Deleted successfully",
	}, nil
}

func (s *BackupStorageService) TestConnection(ctx context.Context, req *pbControlPlane.BackupStorage) (*pbControlPlane.BackupMessageResponse, error) {
	backup := s.protoToEntity(req)
	err := s.service.TestConnection(ctx, backup)
	if err != nil {
		return &pbControlPlane.BackupMessageResponse{
			Status:  pbControlPlane.BackupStatusCode_BACKUP_BAD_REQUEST,
			Message: err.Error(),
		}, nil
	}
	return &pbControlPlane.BackupMessageResponse{
		Status:  pbControlPlane.BackupStatusCode_BACKUP_OK,
		Message: "Connection successful",
	}, nil
}

func (s *BackupStorageService) protoToEntity(pb *pbControlPlane.BackupStorage) *entity.BackupStorage {
	return &entity.BackupStorage{
		Model: model.Model{
			Id: pb.Id,
		},
		Name:      pb.Name,
		Endpoint:  pb.Endpoint,
		Region:    pb.Region,
		Bucket:    pb.Bucket,
		AccessKey: pb.AccessKey,
		SecretKey: pb.SecretKey,
		Status:    pb.Status,
	}
}

func (s *BackupStorageService) entityToProto(e *entity.BackupStorage) *pbControlPlane.BackupStorage {
	return &pbControlPlane.BackupStorage{
		Id:           e.Id,
		Name:         e.Name,
		Endpoint:     e.Endpoint,
		Region:       e.Region,
		Bucket:       e.Bucket,
		AccessKey:    e.AccessKey,
		SecretKey:    e.SecretKey,
		Status:       e.Status,
		ErrorMessage: e.ErrorMessage,
	}
}
