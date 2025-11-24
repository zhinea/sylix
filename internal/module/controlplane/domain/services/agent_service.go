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
	commonWorkflow "github.com/zhinea/sylix/internal/common/workflow"
	pbAgent "github.com/zhinea/sylix/internal/infra/proto/agent"
	"github.com/zhinea/sylix/internal/module/controlplane/domain/repository"
	"github.com/zhinea/sylix/internal/module/controlplane/domain/workflow"
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
			tlsConfig.ServerName = server.IpAddress
			tlsConfig.InsecureSkipVerify = false
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

	remoteConfigPath := "/etc/sylix-agent/config.yaml"
	if err := client.CopyFile(tmpFile, remoteConfigPath); err != nil {
		return fmt.Errorf("failed to copy config file to server: %w", err)
	}

	// Restart agent service
	if _, err := client.RunCommand("systemctl restart sylix-agent"); err != nil {
		return fmt.Errorf("failed to restart agent service: %w", err)
	}

	return nil
}

func (s *AgentService) SyncStorage(ctx context.Context, server *entity.Server, storages []*entity.BackupStorage) error {
	client, err := util.NewSSHClient(server.IpAddress, server.Port, server.Credential.Username, server.Credential.Password, server.Credential.SSHKey)
	if err != nil {
		return err
	}
	defer client.Close()

	// Read existing config
	remoteConfigPath := "/etc/sylix-agent/config.yaml"
	output, err := client.RunCommand("cat " + remoteConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read remote config: %w", err)
	}

	var agentConfig config.AgentConfig
	if err := yaml.Unmarshal([]byte(output), &agentConfig); err != nil {
		return fmt.Errorf("failed to parse remote config: %w", err)
	}

	// Update storage
	var storageConfigs []config.StorageConfig
	for _, st := range storages {
		storageConfigs = append(storageConfigs, config.StorageConfig{
			ID:        st.Id,
			Name:      st.Name,
			Endpoint:  st.Endpoint,
			Region:    st.Region,
			Bucket:    st.Bucket,
			AccessKey: st.AccessKey,
			SecretKey: st.SecretKey,
		})
	}
	agentConfig.Storage = storageConfigs

	// Marshal back to YAML
	newConfigBytes, err := yaml.Marshal(&agentConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal new config: %w", err)
	}

	// Write to temp file
	tmpFile := fmt.Sprintf("agent_config_storage_%s.yaml", server.Id)
	if err := os.WriteFile(tmpFile, newConfigBytes, 0644); err != nil {
		return fmt.Errorf("failed to write temp config file: %w", err)
	}
	defer os.Remove(tmpFile)

	// Copy to server
	if err := client.CopyFile(tmpFile, remoteConfigPath); err != nil {
		return fmt.Errorf("failed to copy config file to server: %w", err)
	}

	// Restart agent
	if _, err := client.RunCommand("systemctl restart sylix-agent"); err != nil {
		return fmt.Errorf("failed to restart agent service: %w", err)
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
	// Setup logger
	logDir := fmt.Sprintf("logs/servers/%s", server.Id)
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

	// Helper to write with timestamp
	logMsg := func(msg string) {
		timestamp := time.Now().Format(time.RFC3339)
		fmt.Fprintf(l, "[%s] %s\n", timestamp, msg)
	}

	s.updateAgentStatus(ctx, server.Id, entity.AgentStatusInstalling)
	logMsg("Starting installation...")

	client, err := util.NewSSHClient(server.IpAddress, server.Port, server.Credential.Username, server.Credential.Password, server.Credential.SSHKey)
	if err != nil {
		logMsg(fmt.Sprintf("Failed to connect via SSH: %v", err))
		s.updateAgentStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	}
	defer client.Close()

	// Prepare dynamic data
	version := os.Getenv("SYLIX_VERSION")
	if version == "" {
		version = common.Version
	}
	if version == "0.0.0-dev" {
		version = "0.1.1"
	}
	downloadURL := fmt.Sprintf("https://github.com/zhinea/sylix/releases/download/v%s/agent", version)
	// remoteBinaryPath := "/usr/local/bin/sylix-agent"

	// Generate Certs
	serverCert, serverKey, err := util.GenerateSelfSignedCert(server.IpAddress)
	if err != nil {
		logMsg(fmt.Sprintf("Failed to generate server cert: %v", err))
		s.updateAgentStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	}
	// Save Cert and Key to server entity
	server.Agent.Cert = string(serverCert)
	server.Agent.Key = string(serverKey)
	if _, err := s.repo.Update(ctx, server); err != nil {
		logMsg(fmt.Sprintf("Failed to save Certs to DB: %v", err))
		s.updateAgentStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	}

	// Config Content
	configContent := fmt.Sprintf(`server:
  port: %d
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
`, server.Agent.Port)

	// Define Workflow
	wf := workflow.NewAgentInstallWorkflow(workflow.AgentInstallParams{
		DownloadURL:   downloadURL,
		ServerCert:    string(serverCert),
		ServerKey:     string(serverKey),
		ConfigContent: configContent,
	})

	// Run Workflow
	engine := commonWorkflow.NewEngine(client, l, logMsg)
	if err := engine.Run(ctx, wf); err != nil {
		logMsg(fmt.Sprintf("Installation failed: %v", err))
		s.updateAgentStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	}

	s.updateAgentStatus(ctx, server.Id, entity.AgentStatusSuccess)
	logMsg("Agent installed successfully.")
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

func (s *AgentService) CreateDatabase(ctx context.Context, server *entity.Server, req *pbAgent.CreateDatabaseRequest) (*pbAgent.CreateDatabaseResponse, error) {
	port := server.Agent.Port
	if port == 0 {
		port = 8083
	}
	target := fmt.Sprintf("%s:%d", server.IpAddress, port)

	logger.Log.Debug("Connecting to agent for CreateDatabase", zap.String("target", target))

	var opts []grpc.DialOption

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	if server.Agent.Cert != "" {
		cp := x509.NewCertPool()
		if cp.AppendCertsFromPEM([]byte(server.Agent.Cert)) {
			tlsConfig.RootCAs = cp
			tlsConfig.ServerName = server.IpAddress
			tlsConfig.InsecureSkipVerify = false
		}
	}
	creds := credentials.NewTLS(tlsConfig)
	opts = append(opts, grpc.WithTransportCredentials(creds))

	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		logger.Log.Error("Failed to dial agent", zap.Error(err), zap.String("target", target))
		return nil, err
	}
	defer conn.Close()

	client := pbAgent.NewAgentClient(conn)
	resp, err := client.CreateDatabase(ctx, req)
	if err != nil {
		logger.Log.Error("Agent RPC CreateDatabase failed", zap.Error(err))
		return nil, err
	}
	return resp, nil
}
