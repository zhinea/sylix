package entity

import "github.com/zhinea/sylix/internal/common/model"

type BackupStorage struct {
	model.Model
	Name         string    `json:"name"`
	Endpoint     string    `json:"endpoint"`
	Region       string    `json:"region"`
	Bucket       string    `json:"bucket"`
	AccessKey    string    `json:"access_key"`
	SecretKey    string    `json:"secret_key"`
	Status       string    `json:"status"`
	ErrorMessage string    `json:"error_message"`
	Servers      []*Server `json:"servers" gorm:"many2many:server_backup_storages;"`
	ServerIDs    []string  `json:"server_ids" gorm:"-"`
}
