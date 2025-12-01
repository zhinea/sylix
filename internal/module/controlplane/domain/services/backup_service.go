package services

import (
	"context"
	"errors"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/zhinea/sylix/internal/module/controlplane/domain/repository"
	"github.com/zhinea/sylix/internal/module/controlplane/entity"
)

type BackupService struct {
	repo       repository.BackupStorageRepository
	serverRepo repository.ServerRepository
}

func NewBackupService(repo repository.BackupStorageRepository, serverRepo repository.ServerRepository) *BackupService {
	return &BackupService{
		repo:       repo,
		serverRepo: serverRepo,
	}
}

func (s *BackupService) Create(ctx context.Context, backup *entity.BackupStorage) (*entity.BackupStorage, error) {
	if err := s.TestConnection(ctx, backup); err != nil {
		return nil, err
	}

	// Ignore ServerIDs on creation, they will be added via Update if needed.
	// This ensures we don't sync to agents on creation.

	backup.Status = "CONNECTED"
	backup.ErrorMessage = ""
	createdBackup, err := s.repo.Create(ctx, backup)
	if err != nil {
		return nil, err
	}

	return createdBackup, nil
}

func (s *BackupService) GetByID(ctx context.Context, id string) (*entity.BackupStorage, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *BackupService) GetAll(ctx context.Context) ([]*entity.BackupStorage, error) {
	return s.repo.GetAll(ctx)
}

func (s *BackupService) Update(ctx context.Context, backup *entity.BackupStorage) (*entity.BackupStorage, error) {
	if err := s.TestConnection(ctx, backup); err != nil {
		return nil, err
	}

	oldBackup, err := s.repo.GetByID(ctx, backup.Id)
	if err != nil {
		return nil, err
	}

	if len(backup.ServerIDs) > 0 {
		servers, err := s.fetchServers(ctx, backup.ServerIDs)
		if err != nil {
			return nil, err
		}
		backup.Servers = servers
	} else {
		backup.Servers = []*entity.Server{}
	}

	backup.Status = "CONNECTED"
	backup.ErrorMessage = ""
	updatedBackup, err := s.repo.Update(ctx, backup)
	if err != nil {
		return nil, err
	}

	affectedServerIDs := make(map[string]bool)
	for _, s := range oldBackup.Servers {
		affectedServerIDs[s.Id] = true
	}
	for _, s := range updatedBackup.Servers {
		affectedServerIDs[s.Id] = true
	}

	return updatedBackup, nil
}

func (s *BackupService) Delete(ctx context.Context, id string) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	return nil
}

func (s *BackupService) fetchServers(ctx context.Context, ids []string) ([]*entity.Server, error) {
	var servers []*entity.Server
	for _, id := range ids {
		server, err := s.serverRepo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		servers = append(servers, server)
	}
	return servers, nil
}

func (s *BackupService) TestConnection(ctx context.Context, backup *entity.BackupStorage) error {
	endpoint := backup.Endpoint
	secure := true
	if strings.HasPrefix(endpoint, "http://") {
		endpoint = strings.TrimPrefix(endpoint, "http://")
		secure = false
	} else if strings.HasPrefix(endpoint, "https://") {
		endpoint = strings.TrimPrefix(endpoint, "https://")
		secure = true
	}

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(backup.AccessKey, backup.SecretKey, ""),
		Secure: secure,
		Region: backup.Region,
	})
	if err != nil {
		return err
	}

	exists, err := minioClient.BucketExists(ctx, backup.Bucket)
	if err != nil {
		return err
	}

	if !exists {
		return errors.New("bucket does not exist")
	}

	return nil
}
