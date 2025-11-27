package entity

import "time"

type NodeType string

const (
	Compute       NodeType = "compute"
	PageServer    NodeType = "pageserver"
	SafeKeeper    NodeType = "safekeeper"
	StorageBroker NodeType = "storage_broker"
)

type Node struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Type            NodeType  `json:"type"`
	PriorityStartup int32     `json:"priority_startup"`
	Fields          string    `json:"fields"`  // JSON string
	Imports         string    `json:"imports"` // JSON string
	Exports         string    `json:"exports"` // JSON string
	ServerID        string    `json:"server_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
