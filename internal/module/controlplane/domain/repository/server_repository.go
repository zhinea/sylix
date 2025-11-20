package repository

import (
	"context"

	"github.com/zhinea/sylix/internal/module/controlplane/entity"
)

type ServerRepository interface {
	Create(ctx context.Context, server *entity.Server) (*entity.Server, error)
	GetByID(ctx context.Context, id string) (*entity.Server, error)
	GetAll(ctx context.Context) ([]*entity.Server, error)
	Update(ctx context.Context, server *entity.Server) (*entity.Server, error)
	Delete(ctx context.Context, id string) error
}
