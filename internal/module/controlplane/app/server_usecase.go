package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/zhinea/sylix/internal/common"
	"github.com/zhinea/sylix/internal/common/logger"
	"github.com/zhinea/sylix/internal/common/util"
	"github.com/zhinea/sylix/internal/module/controlplane/domain/repository"
	"github.com/zhinea/sylix/internal/module/controlplane/entity"
	"go.uber.org/zap"
	"gopkg.in/natefinch/lumberjack.v2"
)

type ServerUseCase struct {
	repo repository.ServerRepository
}

func NewServerUseCase(repo repository.ServerRepository) *ServerUseCase {
	return &ServerUseCase{
		repo: repo,
	}
}

func (uc *ServerUseCase) Create(ctx context.Context, server *entity.Server) (*entity.Server, error) {
	// Set initial status
	server.Status = entity.ServerStatusDisconnected

	// Check connection before creating
	if err := uc.CheckConnection(server); err == nil {
		server.Status = entity.ServerStatusConnected
	} else {
		server.Status = entity.ServerStatusDisconnected
		logger.Log.Warn("Failed to connect to server during creation", zap.Error(err), zap.String("ip", server.IpAddress))
	}

	return uc.repo.Create(ctx, server)
}

func (uc *ServerUseCase) Get(ctx context.Context, id string) (*entity.Server, error) {
	server, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return server, nil
}

func (uc *ServerUseCase) GetAll(ctx context.Context) ([]*entity.Server, error) {
	return uc.repo.GetAll(ctx)
}

func (uc *ServerUseCase) Update(ctx context.Context, server *entity.Server) (*entity.Server, error) {
	// Check connection before updating
	if err := uc.CheckConnection(server); err == nil {
		server.Status = entity.ServerStatusConnected
	} else {
		server.Status = entity.ServerStatusDisconnected
		logger.Log.Warn("Failed to connect to server during update", zap.Error(err), zap.String("ip", server.IpAddress))
	}

	return uc.repo.Update(ctx, server)
}

func (uc *ServerUseCase) RetryConnection(ctx context.Context, id string) (*entity.Server, error) {
	server, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := uc.CheckConnection(server); err == nil {
		server.Status = entity.ServerStatusConnected
	} else {
		server.Status = entity.ServerStatusDisconnected
		logger.Log.Warn("Failed to connect to server during retry", zap.Error(err), zap.String("ip", server.IpAddress))
	}

	return uc.repo.Update(ctx, server)
}

func (uc *ServerUseCase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *ServerUseCase) CheckConnection(server *entity.Server) error {
	client, err := util.NewSSHClient(server.IpAddress, server.Port, server.Credential.Username, server.Credential.Password, server.Credential.SSHKey)
	if err != nil {
		return err
	}
	defer client.Close()

	log.Println(server)

	// Try to run a simple command
	_, err = client.RunCommand("echo 'hello'")
	return err
}

func (uc *ServerUseCase) InstallAgent(ctx context.Context, serverID string) error {
	server, err := uc.repo.GetByID(ctx, serverID)
	if err != nil {
		return fmt.Errorf("failed to get server: %w", err)
	}

	if server.Status != entity.ServerStatusConnected {
		return fmt.Errorf("server must be connected to install agent")
	}

	// Start async installation
	go uc.runAgentInstallation(context.Background(), server)

	return nil
}

func (uc *ServerUseCase) runAgentInstallation(ctx context.Context, server *entity.Server) {
	uc.updateAgentStatus(ctx, server.Id, entity.AgentStatusInstalling)
	uc.appendAgentLog(ctx, server.Id, "Starting installation...")

	client, err := util.NewSSHClient(server.IpAddress, server.Port, server.Credential.Username, server.Credential.Password, server.Credential.SSHKey)
	if err != nil {
		uc.appendAgentLog(ctx, server.Id, fmt.Sprintf("Failed to connect via SSH: %v", err))
		uc.updateAgentStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	}
	defer client.Close()

	// 1. Download Agent Binary
	uc.appendAgentLog(ctx, server.Id, "Downloading agent binary...")
	remoteBinaryPath := "/usr/local/bin/sylix-agent"

	version := common.Version
	if version == "0.0.0-dev" {
		// Fallback for development if version is not injected
		version = "0.1.1"
	}

	downloadURL := fmt.Sprintf("https://github.com/zhinea/sylix/releases/download/v%s/agent", version)
	downloadCmd := fmt.Sprintf("curl -L -f -o %s %s || wget -O %s %s", remoteBinaryPath, downloadURL, remoteBinaryPath, downloadURL)

	if _, err := client.RunCommand(downloadCmd); err != nil {
		uc.appendAgentLog(ctx, server.Id, fmt.Sprintf("Failed to download agent binary from %s: %v", downloadURL, err))
		uc.updateAgentStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	}

	if _, err := client.RunCommand("chmod +x " + remoteBinaryPath); err != nil {
		uc.appendAgentLog(ctx, server.Id, fmt.Sprintf("Failed to make agent executable: %v", err))
		uc.updateAgentStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	}

	// 2. Create Systemd Service
	uc.appendAgentLog(ctx, server.Id, "Creating systemd service...")
	serviceContent := `[Unit]
Description=Sylix Agent
After=network.target

[Service]
ExecStart=/usr/local/bin/sylix-agent
Restart=always
User=root

[Install]
WantedBy=multi-user.target
`
	tmpServiceFile := "sylix-agent.service"
	if err := os.WriteFile(tmpServiceFile, []byte(serviceContent), 0644); err != nil {
		uc.appendAgentLog(ctx, server.Id, fmt.Sprintf("Failed to create temp service file: %v", err))
		uc.updateAgentStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	}
	defer os.Remove(tmpServiceFile)

	remoteServicePath := "/etc/systemd/system/sylix-agent.service"
	if err := client.CopyFile(tmpServiceFile, remoteServicePath); err != nil {
		uc.appendAgentLog(ctx, server.Id, fmt.Sprintf("Failed to copy service file: %v", err))
		uc.updateAgentStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	}

	// 3. Start Service
	uc.appendAgentLog(ctx, server.Id, "Starting service...")
	commands := []string{
		"systemctl daemon-reload",
		"systemctl enable sylix-agent",
		"systemctl restart sylix-agent",
	}

	for _, cmd := range commands {
		if _, err := client.RunCommand(cmd); err != nil {
			uc.appendAgentLog(ctx, server.Id, fmt.Sprintf("Failed to run command %s: %v", cmd, err))
			uc.updateAgentStatus(ctx, server.Id, entity.AgentStatusFailed)
			return
		}
	}

	uc.updateAgentStatus(ctx, server.Id, entity.AgentStatusSuccess)
	uc.appendAgentLog(ctx, server.Id, "Agent installed successfully.")
}

func (uc *ServerUseCase) updateAgentStatus(ctx context.Context, serverID string, status int) {
	server, err := uc.repo.GetByID(ctx, serverID)
	if err != nil {
		logger.Log.Error("Failed to get server for status update", zap.Error(err))
		return
	}
	server.AgentStatus = status
	uc.repo.Update(ctx, server)
}

func (uc *ServerUseCase) appendAgentLog(ctx context.Context, serverID string, logMsg string) {
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
