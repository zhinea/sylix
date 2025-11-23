package services

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/zhinea/sylix/internal/common"
	"github.com/zhinea/sylix/internal/common/config"
	"github.com/zhinea/sylix/internal/common/logger"
	"github.com/zhinea/sylix/internal/common/util"
	pbAgent "github.com/zhinea/sylix/internal/infra/proto/agent"
	"github.com/zhinea/sylix/internal/module/controlplane/domain/repository"
	"github.com/zhinea/sylix/internal/module/controlplane/entity"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v3"
)

type AgentService struct {
	repo repository.ServerRepository
}

func NewAgentService(repo repository.ServerRepository) *AgentService {
	return &AgentService{
		repo: repo,
	}
}

func (s *AgentService) CheckConnection(server *entity.Server) error {
	client, err := util.NewSSHClient(server.IpAddress, server.Port, server.Credential.Username, server.Credential.Password, server.Credential.SSHKey)
	if err != nil {
		return err
	}
	defer client.Close()
	return nil
}

func (s *AgentService) GetAgentConfig(ctx context.Context, server *entity.Server) (string, string, error) {
	port := server.Agent.Port
	if port == 0 {
		port = 8083
	}
	target := fmt.Sprintf("%s:%d", server.IpAddress, port)

	var opts []grpc.DialOption

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	if server.Agent.Cert != "" {
		cp := x509.NewCertPool()
		if cp.AppendCertsFromPEM([]byte(server.Agent.Cert)) {
			tlsConfig.RootCAs = cp
		}
	}
	creds := credentials.NewTLS(tlsConfig)
	opts = append(opts, grpc.WithTransportCredentials(creds))

	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return "", "", err
	}
	defer conn.Close()

	client := pbAgent.NewAgentClient(conn)
	resp, err := client.GetConfig(ctx, &pbAgent.GetConfigRequest{})
	if err != nil {
		return "", "", err
	}

	return resp.Config, resp.Timezone, nil
}

func (s *AgentService) Install(ctx context.Context, server *entity.Server) {
	s.runAgentInstallation(ctx, server)
}

func (s *AgentService) Configure(ctx context.Context, server *entity.Server, configStr string) error {
	client, err := util.NewSSHClient(server.IpAddress, server.Port, server.Credential.Username, server.Credential.Password, server.Credential.SSHKey)
	if err != nil {
		return err
	}
	defer client.Close()

	// Write config to a temporary local file
	tmpFile := fmt.Sprintf("agent_config_%s.yaml", server.Id)
	if err := os.WriteFile(tmpFile, []byte(configStr), 0644); err != nil {
		return fmt.Errorf("failed to write temp config file: %w", err)
	}
	defer os.Remove(tmpFile)

	// Copy to remote server
	remotePath := "/etc/sylix-agent/config.yaml"
	if err := client.CopyFile(tmpFile, remotePath); err != nil {
		return fmt.Errorf("failed to copy config file to server: %w", err)
	}

	// Restart agent
	if _, err := client.RunCommand("systemctl restart sylix-agent"); err != nil {
		return fmt.Errorf("failed to restart agent: %w", err)
	}

	return nil
}

func (s *AgentService) UpdatePort(ctx context.Context, server *entity.Server, port int) error {
	client, err := util.NewSSHClient(server.IpAddress, server.Port, server.Credential.Username, server.Credential.Password, server.Credential.SSHKey)
	if err != nil {
		return err
	}
	defer client.Close()

	// Read remote config
	remotePath := "/etc/sylix-agent/config.yaml"
	out, err := client.RunCommand("cat " + remotePath)
	if err != nil {
		return fmt.Errorf("failed to read remote config: %w", err)
	}

	// Parse config
	var cfg config.AgentConfig
	if err := yaml.Unmarshal([]byte(out), &cfg); err != nil {
		return fmt.Errorf("failed to parse remote config: %w", err)
	}

	// Update port
	cfg.Server.Port = port

	// Marshal config
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	tmpFile := fmt.Sprintf("agent_config_update_%s.yaml", server.Id)
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp config file: %w", err)
	}
	defer os.Remove(tmpFile)

	if err := client.CopyFile(tmpFile, remotePath); err != nil {
		return fmt.Errorf("failed to copy config file to server: %w", err)
	}

	if _, err := client.RunCommand("systemctl restart sylix-agent"); err != nil {
		return fmt.Errorf("failed to restart agent: %w", err)
	}

	// Update agent port in DB
	server.Agent.Port = port
	if _, err := s.repo.Update(ctx, server); err != nil {
		return fmt.Errorf("failed to update server agent port in db: %w", err)
	}

	return nil
}

func (s *AgentService) UpdateTimeZone(ctx context.Context, server *entity.Server, timezone string) error {
	client, err := util.NewSSHClient(server.IpAddress, server.Port, server.Credential.Username, server.Credential.Password, server.Credential.SSHKey)
	if err != nil {
		return err
	}
	defer client.Close()

	// Install chrony if not present
	installCmd := "if ! command -v chronyd >/dev/null 2>&1; then apt-get update && apt-get install -y chrony || yum install -y chrony; fi"
	if _, err := client.RunCommand(installCmd); err != nil {
		return fmt.Errorf("failed to install chrony: %w", err)
	}

	// Set timezone
	tzCmd := fmt.Sprintf("timedatectl set-timezone %s", timezone)
	if _, err := client.RunCommand(tzCmd); err != nil {
		return fmt.Errorf("failed to set timezone: %w", err)
	}

	// Enable NTP
	ntpCmd := "timedatectl set-ntp true"
	if _, err := client.RunCommand(ntpCmd); err != nil {
		return fmt.Errorf("failed to enable NTP: %w", err)
	}

	return nil
}

func (s *AgentService) runAgentInstallation(ctx context.Context, server *entity.Server) {
	s.updateAgentStatus(ctx, server.Id, entity.AgentStatusInstalling)
	s.appendAgentLog(ctx, server.Id, "Starting installation...")

	client, err := util.NewSSHClient(server.IpAddress, server.Port, server.Credential.Username, server.Credential.Password, server.Credential.SSHKey)
	if err != nil {
		s.appendAgentLog(ctx, server.Id, fmt.Sprintf("Failed to connect via SSH: %v", err))
		s.updateAgentStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	}
	defer client.Close()

	// 0. Stop existing service if running
	s.appendAgentLog(ctx, server.Id, "Stopping existing service (if any)...")
	stopCmd := "systemctl stop sylix-agent || true"
	if out, err := client.RunCommand(stopCmd); err != nil {
		s.appendAgentLog(ctx, server.Id, fmt.Sprintf("Warning: Failed to stop service: %v", err))
	} else if out != "" {
		s.appendAgentLog(ctx, server.Id, fmt.Sprintf("Stop service output: %s", out))
	}

	// 1. Download Agent Binary
	s.appendAgentLog(ctx, server.Id, "Downloading agent binary...")
	remoteBinaryPath := "/usr/local/bin/sylix-agent"

	version := os.Getenv("SYLIX_VERSION")
	if version == "" {
		version = common.Version
	}
	if version == "0.0.0-dev" {
		// Fallback for development if version is not injected
		version = "0.1.1"
	}

	downloadURL := fmt.Sprintf("https://github.com/zhinea/sylix/releases/download/v%s/agent", version)
	// Try curl first, then wget. Split commands to get better error reporting.
	downloadCmd := fmt.Sprintf("if command -v curl >/dev/null 2>&1; then curl -L -f -o %s %s; elif command -v wget >/dev/null 2>&1; then wget -O %s %s; else echo 'Error: neither curl nor wget found'; exit 1; fi", remoteBinaryPath, downloadURL, remoteBinaryPath, downloadURL)

	if out, err := client.RunCommand(downloadCmd); err != nil {
		s.appendAgentLog(ctx, server.Id, fmt.Sprintf("Failed to download agent binary from %s: %v", downloadURL, err))
		s.updateAgentStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	} else {
		s.appendAgentLog(ctx, server.Id, fmt.Sprintf("Download output: %s", out))
	}

	if _, err := client.RunCommand("chmod +x " + remoteBinaryPath); err != nil {
		s.appendAgentLog(ctx, server.Id, fmt.Sprintf("Failed to make agent executable: %v", err))
		s.updateAgentStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	}

	// 1.5 Generate and Install Certificates
	s.appendAgentLog(ctx, server.Id, "Generating certificates...")
	caCert, caKey, err := util.GenerateCA()
	if err != nil {
		s.appendAgentLog(ctx, server.Id, fmt.Sprintf("Failed to generate CA: %v", err))
		s.updateAgentStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	}

	serverCert, serverKey, err := util.GenerateCert(caCert, caKey, server.IpAddress)
	if err != nil {
		s.appendAgentLog(ctx, server.Id, fmt.Sprintf("Failed to generate server cert: %v", err))
		s.updateAgentStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	}

	// Save CA to server entity
	server.Agent.Cert = string(caCert)
	if _, err := s.repo.Update(ctx, server); err != nil {
		s.appendAgentLog(ctx, server.Id, fmt.Sprintf("Failed to save CA to DB: %v", err))
		s.updateAgentStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	}

	// Create certs directory
	certsDir := "/etc/sylix-agent/certs"
	if _, err := client.RunCommand("mkdir -p " + certsDir); err != nil {
		s.appendAgentLog(ctx, server.Id, fmt.Sprintf("Failed to create certs dir: %v", err))
		s.updateAgentStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	}

	// Upload certs
	tmpCertFile := fmt.Sprintf("server_cert_%s.pem", server.Id)
	if err := os.WriteFile(tmpCertFile, serverCert, 0644); err != nil {
		s.appendAgentLog(ctx, server.Id, fmt.Sprintf("Failed to write temp cert file: %v", err))
		s.updateAgentStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	}
	defer os.Remove(tmpCertFile)

	if err := client.CopyFile(tmpCertFile, filepath.Join(certsDir, "server.crt")); err != nil {
		s.appendAgentLog(ctx, server.Id, fmt.Sprintf("Failed to copy cert file: %v", err))
		s.updateAgentStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	}

	tmpKeyFile := fmt.Sprintf("server_key_%s.pem", server.Id)
	if err := os.WriteFile(tmpKeyFile, serverKey, 0644); err != nil {
		s.appendAgentLog(ctx, server.Id, fmt.Sprintf("Failed to write temp key file: %v", err))
		s.updateAgentStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	}
	defer os.Remove(tmpKeyFile)

	if err := client.CopyFile(tmpKeyFile, filepath.Join(certsDir, "server.key")); err != nil {
		s.appendAgentLog(ctx, server.Id, fmt.Sprintf("Failed to copy key file: %v", err))
		s.updateAgentStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	}

	// 2. Create Config File
	s.appendAgentLog(ctx, server.Id, "Creating configuration file...")
	configDir := "/etc/sylix-agent"
	if _, err := client.RunCommand("mkdir -p " + configDir); err != nil {
		s.appendAgentLog(ctx, server.Id, fmt.Sprintf("Failed to create config dir: %v", err))
		s.updateAgentStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	}

	defaultConfig := `server:
  port: 8083
  host: "0.0.0.0"
security:
  cert_file: "/etc/sylix-agent/certs/server.crt"
  key_file: "/etc/sylix-agent/certs/server.key"
log:
  level: "info"
  filename: "/etc/sylix-agent/agent.log"
  max_size: 10
  max_backups: 3
  max_age: 28
  compress: true
`
	tmpConfigFile := fmt.Sprintf("agent_config_setup_%s.yaml", server.Id)
	if err := os.WriteFile(tmpConfigFile, []byte(defaultConfig), 0644); err != nil {
		s.appendAgentLog(ctx, server.Id, fmt.Sprintf("Failed to create temp config file: %v", err))
		s.updateAgentStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	}
	defer os.Remove(tmpConfigFile)

	remoteConfigPath := filepath.Join(configDir, "config.yaml")
	if err := client.CopyFile(tmpConfigFile, remoteConfigPath); err != nil {
		s.appendAgentLog(ctx, server.Id, fmt.Sprintf("Failed to copy config file: %v", err))
		s.updateAgentStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	}

	// 3. Create Systemd Service
	s.appendAgentLog(ctx, server.Id, "Creating systemd service...")
	serviceContent := `[Unit]
Description=Sylix Agent
After=network.target

[Service]
ExecStart=/usr/local/bin/sylix-agent -config /etc/sylix-agent/config.yaml
Restart=always
User=root

[Install]
WantedBy=multi-user.target
`
	tmpServiceFile := "sylix-agent.service"
	if err := os.WriteFile(tmpServiceFile, []byte(serviceContent), 0644); err != nil {
		s.appendAgentLog(ctx, server.Id, fmt.Sprintf("Failed to create temp service file: %v", err))
		s.updateAgentStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	}
	defer os.Remove(tmpServiceFile)

	remoteServicePath := "/etc/systemd/system/sylix-agent.service"
	if err := client.CopyFile(tmpServiceFile, remoteServicePath); err != nil {
		s.appendAgentLog(ctx, server.Id, fmt.Sprintf("Failed to copy service file: %v", err))
		s.updateAgentStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	}

	// 3. Start Service
	s.appendAgentLog(ctx, server.Id, "Starting service...")
	commands := []string{
		"systemctl daemon-reload",
		"systemctl enable sylix-agent",
		"systemctl restart sylix-agent",
	}

	for _, cmd := range commands {
		if _, err := client.RunCommand(cmd); err != nil {
			s.appendAgentLog(ctx, server.Id, fmt.Sprintf("Failed to run command %s: %v", cmd, err))
			s.updateAgentStatus(ctx, server.Id, entity.AgentStatusFailed)
			return
		}
	}

	s.updateAgentStatus(ctx, server.Id, entity.AgentStatusSuccess)
	s.appendAgentLog(ctx, server.Id, "Agent installed successfully.")
}

func (s *AgentService) updateAgentStatus(ctx context.Context, serverID string, status int) {
	server, err := s.repo.GetByID(ctx, serverID)
	if err != nil {
		logger.Log.Error("Failed to get server for status update", zap.Error(err))
		return
	}
	server.Agent.Status = status
	s.repo.Update(ctx, server)
}

func (s *AgentService) appendAgentLog(ctx context.Context, serverID string, logMsg string) {
	logDir := fmt.Sprintf("logs/servers/%s", serverID)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		logger.Log.Error("Failed to create log dir", zap.Error(err))
		return
	}

	logFile := filepath.Join(logDir, "setup_agent.log")
	l := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    25, // MB
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}
	defer l.Close()

	msg := fmt.Sprintf("[%s] %s\n", time.Now().Format(time.RFC3339), logMsg)
	if _, err := l.Write([]byte(msg)); err != nil {
		logger.Log.Error("Failed to write agent log", zap.Error(err))
	}
}
