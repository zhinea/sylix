package repository

import (
	"context"

	"github.com/zhinea/sylix/internal/module/controlplane/entity"
	"gorm.io/gorm"
)

type ServerRepositoryImpl struct {
	db *gorm.DB
}

func NewServerRepository(db *gorm.DB) ServerRepository {
	return &ServerRepositoryImpl{
		db: db,
	}
}

func (s *ServerRepositoryImpl) Create(ctx context.Context, server *entity.Server) (*entity.Server, error) {
	if err := s.db.WithContext(ctx).Create(server).Error; err != nil {
		return nil, err
	}
	return server, nil
}

func (s *ServerRepositoryImpl) GetByID(ctx context.Context, id string) (*entity.Server, error) {
	var server entity.Server
	if err := s.db.WithContext(ctx).First(&server, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &server, nil
}

func (s *ServerRepositoryImpl) GetAll(ctx context.Context) ([]*entity.Server, error) {
	var servers []*entity.Server
	if err := s.db.WithContext(ctx).Find(&servers).Error; err != nil {
		return nil, err
	}
	return servers, nil
}

func (s *ServerRepositoryImpl) Update(ctx context.Context, server *entity.Server) (*entity.Server, error) {
	if err := s.db.WithContext(ctx).Save(server).Error; err != nil {
		return nil, err
	}
	return server, nil
}

func (s *ServerRepositoryImpl) Delete(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&entity.Server{}, "id = ?", id).Error
}
