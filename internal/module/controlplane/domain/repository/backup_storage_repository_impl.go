package repository

import (
	"context"

	"github.com/zhinea/sylix/internal/module/controlplane/entity"
	"gorm.io/gorm"
)

type BackupStorageRepositoryImpl struct {
	db *gorm.DB
}

func NewBackupStorageRepository(db *gorm.DB) BackupStorageRepository {
	return &BackupStorageRepositoryImpl{
		db: db,
	}
}

func (r *BackupStorageRepositoryImpl) Create(ctx context.Context, backup *entity.BackupStorage) (*entity.BackupStorage, error) {
	if err := r.db.WithContext(ctx).Create(backup).Error; err != nil {
		return nil, err
	}
	return backup, nil
}

func (r *BackupStorageRepositoryImpl) GetByID(ctx context.Context, id string) (*entity.BackupStorage, error) {
	var backup entity.BackupStorage
	if err := r.db.WithContext(ctx).First(&backup, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &backup, nil
}

func (r *BackupStorageRepositoryImpl) GetAll(ctx context.Context) ([]*entity.BackupStorage, error) {
	var backups []*entity.BackupStorage
	if err := r.db.WithContext(ctx).Find(&backups).Error; err != nil {
		return nil, err
	}
	return backups, nil
}

func (r *BackupStorageRepositoryImpl) Update(ctx context.Context, backup *entity.BackupStorage) (*entity.BackupStorage, error) {
	if err := r.db.WithContext(ctx).Save(backup).Error; err != nil {
		return nil, err
	}
	return backup, nil
}

func (r *BackupStorageRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.BackupStorage{}, "id = ?", id).Error
}
