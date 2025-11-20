package db

import (
	"github.com/zhinea/sylix/internal/module/controlplane/entity"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&entity.Server{},
	)
}
