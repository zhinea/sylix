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
	repo repository.BackupStorageRepository
}

func NewBackupService(repo repository.BackupStorageRepository) *BackupService {
	return &BackupService{
		repo: repo,
	}
}

func (s *BackupService) Create(ctx context.Context, backup *entity.BackupStorage) (*entity.BackupStorage, error) {
	if err := s.TestConnection(ctx, backup); err != nil {
		return nil, err
	}

	backup.Status = "CONNECTED"
	backup.ErrorMessage = ""
	return s.repo.Create(ctx, backup)
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

	backup.Status = "CONNECTED"
	backup.ErrorMessage = ""
	return s.repo.Update(ctx, backup)
}

func (s *BackupService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
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
