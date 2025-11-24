package repository

import (
	"context"

	"github.com/zhinea/sylix/internal/module/controlplane/entity"
)

type DatabaseRepository interface {
	Create(ctx context.Context, database *entity.Database) (*entity.Database, error)
	GetByID(ctx context.Context, id string) (*entity.Database, error)
	GetAll(ctx context.Context) ([]*entity.Database, error)
	GetByServerID(ctx context.Context, serverID string) ([]*entity.Database, error)
	Update(ctx context.Context, database *entity.Database) (*entity.Database, error)
	Delete(ctx context.Context, id string) error
}
