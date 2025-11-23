package db

import (
	"github.com/zhinea/sylix/internal/module/controlplane/entity"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	// Drop agent_logs column if it exists (cleanup)
	if db.Migrator().HasColumn(&entity.Server{}, "agent_logs") {
		if err := db.Migrator().DropColumn(&entity.Server{}, "agent_logs"); err != nil {
			return err
		}
	}

	return db.AutoMigrate(
		&entity.Server{},
	)
}
