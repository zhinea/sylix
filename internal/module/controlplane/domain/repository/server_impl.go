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
	// tx := s.db.WithContext(ctx).Create(server)
	// tx.
	return &entity.Server{}, nil
}

func (s *ServerRepositoryImpl) GetByID(ctx context.Context, id string) (*entity.Server, error) {
	return &entity.Server{}, nil
}

func (s *ServerRepositoryImpl) GetAll(ctx context.Context) ([]*entity.Server, error) {
	return []*entity.Server{}, nil
}

func (s *ServerRepositoryImpl) Update(ctx context.Context, server *entity.Server) (*entity.Server, error) {
	return &entity.Server{}, nil
}

func (s *ServerRepositoryImpl) Delete(ctx context.Context, id string) error {
	return nil
}
