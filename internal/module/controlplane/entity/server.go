package entity

import "github.com/zhinea/sylix/internal/common/model"

type ServerCredential struct {
	Username string  `json:"username"`
	Password *string `json:"password,omitempty"`
	SSHKey   *string `json:"ssh_key,omitempty"`
}

type Server struct {
	model.Model
	Name       string           `json:"name"`
	IpAddress  string           `json:"ip_address"`
	Port       int              `json:"port"`
	Protocol   string           `json:"protocol"`
	Credential ServerCredential `json:"credential" gorm:"embedded;embeddedPrefix:credential_"`
}
