package entity

import "github.com/zhinea/sylix/internal/common/model"

type ServerCredential struct {
	Username string  `json:"username"`
	Password *string `json:"password,omitempty"`
	SSHKey   *string `json:"ssh_key,omitempty"`
}

type ServerAgent struct {
	Port   int    `json:"port"`
	Status int    `json:"status"`
	Logs   string `json:"logs" gorm:"-"`
	Cert   string `json:"-"`
	Key    string `json:"-"`
}

type Server struct {
	model.Model
	Name           string           `json:"name"`
	IpAddress      string           `json:"ip_address"`
	Port           int              `json:"port"`
	Protocol       string           `json:"protocol"`
	Credential     ServerCredential `json:"credential" gorm:"embedded;embeddedPrefix:credential_"`
	Agent          ServerAgent      `json:"agent" gorm:"embedded;embeddedPrefix:agent_"`
	Status         int              `json:"status"`
	BackupStorages []*BackupStorage `json:"backup_storages" gorm:"many2many:server_backup_storages;"`
}

const (
	ServerStatusUnspecified  = 0
	ServerStatusConnected    = 1
	ServerStatusDisconnected = 2
)

const (
	AgentStatusUnspecified     = 0
	AgentStatusInstalling      = 1
	AgentStatusConfiguring     = 2
	AgentStatusFinalizingSetup = 3
	AgentStatusSuccess         = 4
	AgentStatusFailed          = 5
)
