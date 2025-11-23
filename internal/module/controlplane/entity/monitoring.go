package entity

import (
	"time"

	"github.com/zhinea/sylix/internal/common/model"
)

type ServerPing struct {
	model.Model
	ServerID     string `json:"server_id" gorm:"index"`
	ResponseTime int64  `json:"response_time"` // in milliseconds
	Status       string `json:"status"`
	Error        string `json:"error"`
}

type ServerStat struct {
	model.Model
	ServerID            string    `json:"server_id" gorm:"index"`
	AverageResponseTime float64   `json:"average_response_time"`
	MinResponseTime     int64     `json:"min_response_time"`
	MaxResponseTime     int64     `json:"max_response_time"`
	PingCount           int64     `json:"ping_count"`
	SuccessRate         float64   `json:"success_rate"`
	Timestamp           time.Time `json:"timestamp"`
}

type ServerAccident struct {
	model.Model
	ServerID     string `json:"server_id" gorm:"index"`
	ResponseTime int64  `json:"response_time"`
	Error        string `json:"error"`
	Details      string `json:"details"`
	Resolved     bool   `json:"resolved"`
}
