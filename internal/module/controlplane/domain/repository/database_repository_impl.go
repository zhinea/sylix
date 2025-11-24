package repository

import (
	"context"

	"github.com/zhinea/sylix/internal/module/controlplane/entity"
	"gorm.io/gorm"
)

type DatabaseRepositoryImpl struct {
	db *gorm.DB
}

func NewDatabaseRepository(db *gorm.DB) DatabaseRepository {
	return &DatabaseRepositoryImpl{
		db: db,
	}
}

func (r *DatabaseRepositoryImpl) Create(ctx context.Context, database *entity.Database) (*entity.Database, error) {
	if err := r.db.WithContext(ctx).Create(database).Error; err != nil {
		return nil, err
	}
	return database, nil
}

func (r *DatabaseRepositoryImpl) GetByID(ctx context.Context, id string) (*entity.Database, error) {
	var database entity.Database
	if err := r.db.WithContext(ctx).Preload("Server").First(&database, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &database, nil
}

func (r *DatabaseRepositoryImpl) GetAll(ctx context.Context) ([]*entity.Database, error) {
	var databases []*entity.Database
	if err := r.db.WithContext(ctx).Preload("Server").Find(&databases).Error; err != nil {
		return nil, err
	}
	return databases, nil
}

func (r *DatabaseRepositoryImpl) GetByServerID(ctx context.Context, serverID string) ([]*entity.Database, error) {
	var databases []*entity.Database
	if err := r.db.WithContext(ctx).Where("server_id = ?", serverID).Find(&databases).Error; err != nil {
		return nil, err
	}
	return databases, nil
}

func (r *DatabaseRepositoryImpl) Update(ctx context.Context, database *entity.Database) (*entity.Database, error) {
	if err := r.db.WithContext(ctx).Save(database).Error; err != nil {
		return nil, err
	}
	return database, nil
}

func (r *DatabaseRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Database{}, "id = ?", id).Error
}
