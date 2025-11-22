package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/zhinea/sylix/internal/common/util"
	"github.com/zhinea/sylix/internal/module/controlplane/domain/repository"
	"github.com/zhinea/sylix/internal/module/controlplane/entity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	// Check connection before creating
	if err := uc.CheckConnection(server); err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	return uc.repo.Create(ctx, server)
}

func (uc *ServerUseCase) Get(ctx context.Context, id string) (*entity.Server, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *ServerUseCase) GetAll(ctx context.Context) ([]*entity.Server, error) {
	return uc.repo.GetAll(ctx)
}

func (uc *ServerUseCase) Update(ctx context.Context, server *entity.Server) (*entity.Server, error) {
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

	// Try to run a simple command
	_, err = client.RunCommand("echo 'hello'")
	return err
}

func (uc *ServerUseCase) InstallAgent(ctx context.Context, serverID string) error {
	server, err := uc.repo.GetByID(ctx, serverID)
	if err != nil {
		return fmt.Errorf("failed to get server: %w", err)
	}

	client, err := util.NewSSHClient(server.IpAddress, server.Port, server.Credential.Username, server.Credential.Password, server.Credential.SSHKey)
	if err != nil {
		return fmt.Errorf("failed to connect via SSH: %w", err)
	}
	defer client.Close()

	// 1. Copy Agent Binary
	// Assuming the binary is in the current working directory under bin/agent
	// In a real deployment, this path should be configured or the binary embedded.
	localBinaryPath := "bin/agent"
	remoteBinaryPath := "/usr/local/bin/sylix-agent"

	// Check if local binary exists
	if _, err := os.Stat(localBinaryPath); os.IsNotExist(err) {
		return fmt.Errorf("agent binary not found at %s", localBinaryPath)
	}

	if err := client.CopyFile(localBinaryPath, remoteBinaryPath); err != nil {
		return fmt.Errorf("failed to copy agent binary: %w", err)
	}

	// Make it executable
	if _, err := client.RunCommand("chmod +x " + remoteBinaryPath); err != nil {
		return fmt.Errorf("failed to make agent executable: %w", err)
	}

	// 2. Create Systemd Service
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
	// Write service file to a temporary file locally, then copy it
	tmpServiceFile := "sylix-agent.service"
	if err := os.WriteFile(tmpServiceFile, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to create temp service file: %w", err)
	}
	defer os.Remove(tmpServiceFile)

	remoteServicePath := "/etc/systemd/system/sylix-agent.service"
	if err := client.CopyFile(tmpServiceFile, remoteServicePath); err != nil {
		return fmt.Errorf("failed to copy service file: %w", err)
	}

	// 3. Start Service
	commands := []string{
		"systemctl daemon-reload",
		"systemctl enable sylix-agent",
		"systemctl restart sylix-agent",
	}

	for _, cmd := range commands {
		if _, err := client.RunCommand(cmd); err != nil {
			return fmt.Errorf("failed to run command %s: %w", cmd, err)
		}
	}

	// 4. Check Agent Connection via gRPC
	// Wait a bit for the agent to start
	time.Sleep(2 * time.Second)

	// Assuming agent runs on port 8083 (needs to be consistent with agent config)
	agentAddr := fmt.Sprintf("%s:8083", server.IpAddress)
	conn, err := grpc.DialContext(ctx, agentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
	if err != nil {
		return fmt.Errorf("failed to connect to agent at %s: %w", agentAddr, err)
	}
	defer conn.Close()

	// Ideally we should call a Ping method on the agent, but connection success is a good start.
	// If we had an AgentService client, we would use it here.

	return nil
}
