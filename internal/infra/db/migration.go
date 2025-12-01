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

	if err := db.AutoMigrate(
		&entity.Server{},
		&entity.ServerPing{},
		&entity.ServerStat{},

		&entity.BackupStorage{},
	); err != nil {
		return err
	}

	// Migrate credential_ca_cert to agent_cert
	if db.Migrator().HasColumn(&entity.Server{}, "credential_ca_cert") {
		if err := db.Exec("UPDATE servers SET agent_cert = credential_ca_cert WHERE (agent_cert IS NULL OR agent_cert = '') AND credential_ca_cert IS NOT NULL").Error; err != nil {
			return err
		}
		// Optional: Drop the old column
		// if err := db.Migrator().DropColumn(&entity.Server{}, "credential_ca_cert"); err != nil {
		// 	return err
		// }
	}

	return nil
}
