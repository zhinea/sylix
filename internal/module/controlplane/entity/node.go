package entity

import (
	"time"

	"gorm.io/datatypes"
)

type NodeGraph struct {
	ID        string         `gorm:"primaryKey" json:"id"`
	Name      string         `json:"name"`
	Nodes     datatypes.JSON `json:"nodes"` // Stored as JSON
	Edges     datatypes.JSON `json:"edges"` // Stored as JSON
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// Helper structs for JSON unmarshalling (not stored directly in DB columns)
type Node struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Label    string       `json:"label"`
	Position NodePosition `json:"position"`
	Data     NodeData     `json:"data"`
}

type NodePosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type NodeData struct {
	ServerID        string `json:"server_id"`
	PgVersion       string `json:"pg_version,omitempty"`
	PgPort          int32  `json:"pg_port,omitempty"`
	ExposeInternet  bool   `json:"expose_internet,omitempty"`
	BackupStorageID string `json:"backup_storage_id,omitempty"`
}

type Edge struct {
	ID           string `json:"id"`
	Source       string `json:"source"`
	Target       string `json:"target"`
	SourceHandle string `json:"source_handle"`
	TargetHandle string `json:"target_handle"`
}
