package repository

import (
	"context"

	"github.com/zhinea/sylix/internal/module/controlplane/entity"
)

type BackupStorageRepository interface {
	Create(ctx context.Context, backup *entity.BackupStorage) (*entity.BackupStorage, error)
	GetByID(ctx context.Context, id string) (*entity.BackupStorage, error)
	GetByServerID(ctx context.Context, serverID string) ([]*entity.BackupStorage, error)
	GetAll(ctx context.Context) ([]*entity.BackupStorage, error)
	Update(ctx context.Context, backup *entity.BackupStorage) (*entity.BackupStorage, error)
	Delete(ctx context.Context, id string) error
}
