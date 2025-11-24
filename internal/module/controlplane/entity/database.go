package entity

import "github.com/zhinea/sylix/internal/common/model"

type Database struct {
	model.Model
	Name        string  `json:"name"`
	User        string  `json:"user"`
	Password    string  `json:"password"`
	DbName      string  `json:"db_name"`
	Branch      string  `json:"branch"` // Default: "main"
	ServerID    string  `json:"server_id"`
	Server      *Server `json:"server" gorm:"foreignKey:ServerID"`
	Status      string  `json:"status"` // e.g., CREATING, RUNNING, STOPPED, ERROR
	ContainerID string  `json:"container_id"`
	Port        int     `json:"port"`
}

const (
	DatabaseStatusCreating = "CREATING"
	DatabaseStatusRunning  = "RUNNING"
	DatabaseStatusStopped  = "STOPPED"
	DatabaseStatusError    = "ERROR"
)
