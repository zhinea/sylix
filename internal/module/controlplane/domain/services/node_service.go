package services

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/zhinea/sylix/internal/common/logger"
	"github.com/zhinea/sylix/internal/common/util"
	"github.com/zhinea/sylix/internal/module/controlplane/domain/repository"
	"github.com/zhinea/sylix/internal/module/controlplane/entity"
	"go.uber.org/zap"
)

type NodeService struct {
	repo repository.ServerRepository
}

func NewNodeService(repo repository.ServerRepository) *NodeService {
	return &NodeService{
		repo: repo,
	}
}

func (s *NodeService) CheckConnection(server *entity.Server) error {
	client, err := util.NewSSHClient(server.IpAddress, server.Port, server.Credential.Username, server.Credential.Password, server.Credential.SSHKey)
	if err != nil {
		return err
	}
	defer client.Close()

	// Simple check
	if _, err := client.RunCommand("echo 'ping'"); err != nil {
		return fmt.Errorf("failed to run ping command: %w", err)
	}

	return nil
}

// Install provisions the node with Docker and WireGuard
func (s *NodeService) Install(ctx context.Context, server *entity.Server) {
	go s.runProvisioning(ctx, server)
}

func (s *NodeService) runProvisioning(ctx context.Context, server *entity.Server) {
	logger.Log.Info("Starting node provisioning", zap.String("server_id", server.Id), zap.String("ip", server.IpAddress))

	s.updateStatus(ctx, server.Id, entity.AgentStatusInstalling) // reusing AgentStatus for now

	// Create logs directory
	logDir := fmt.Sprintf("logs/servers/%s", server.Id)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		logger.Log.Error("Failed to create log directory", zap.Error(err))
		// Continue anyway, but logging might fail
	}

	logFile, err := os.OpenFile(filepath.Join(logDir, "provisioning.log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		logger.Log.Error("Failed to create log file", zap.Error(err))
		// Continue anyway
	} else {
		defer logFile.Close()
	}

	// Helper to write to log file
	writeLog := func(msg string) {
		if logFile != nil {
			timestamp := time.Now().Format(time.RFC3339)
			logFile.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, msg))
		}
	}

	client, err := util.NewSSHClient(server.IpAddress, server.Port, server.Credential.Username, server.Credential.Password, server.Credential.SSHKey)
	if err != nil {
		logger.Log.Error("Failed to connect via SSH", zap.Error(err))
		writeLog(fmt.Sprintf("Failed to connect via SSH: %v", err))
		s.updateStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	}
	defer client.Close()

	// Helper to run command and stream logs
	runCmd := func(cmd string) error {
		writeLog(fmt.Sprintf("Running command: %s", cmd))
		logger.Log.Info("Running command", zap.String("cmd", cmd), zap.String("server_id", server.Id))

		if logFile != nil {
			// Create a multi-writer if we want to log to stdout too, but for now just file is requested for "logs"
			// We use the logFile directly for stdout/stderr of the command
			// But we need to be careful not to mix interleaved writes if we use goroutines (RunCommandStream doesn't use goroutines for writers)
			return client.RunCommandStream(cmd, logFile, logFile)
		}

		// Fallback if log file failed
		_, err := client.RunCommand(cmd)
		return err
	}

	// 1. Update & Install Dependencies
	writeLog("Installing dependencies...")
	cmds := []string{
		"apt-get update",
		"apt-get install -y curl wget gnupg lsb-release wireguard", // Install WireGuard directly
	}

	for _, cmd := range cmds {
		if err := runCmd(cmd); err != nil {
			logger.Log.Error("Failed to run command", zap.String("cmd", cmd), zap.Error(err))
			writeLog(fmt.Sprintf("Failed to run command: %v", err))
			s.updateStatus(ctx, server.Id, entity.AgentStatusFailed)
			return
		}
	}

	// 2. Install Docker (if not present)
	writeLog("Checking Docker...")
	if _, err := client.RunCommand("docker --version"); err != nil {
		writeLog("Installing Docker...")
		installDockerCmd := "curl -fsSL https://get.docker.com | sh"
		if err := runCmd(installDockerCmd); err != nil {
			logger.Log.Error("Failed to install Docker", zap.Error(err))
			writeLog(fmt.Sprintf("Failed to install Docker: %v", err))
			s.updateStatus(ctx, server.Id, entity.AgentStatusFailed)
			return
		}
	}

	// 2.5 Configure Docker Daemon (MTU)
	writeLog("Configuring Docker Daemon...")
	if err := s.configureDockerDaemon(client); err != nil {
		logger.Log.Error("Failed to configure Docker Daemon", zap.Error(err))
		writeLog(fmt.Sprintf("Failed to configure Docker Daemon: %v", err))
		s.updateStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	}

	// 3. Setup WireGuard
	writeLog("Setting up WireGuard...")
	if err := s.setupWireGuard(ctx, client, server); err != nil {
		logger.Log.Error("Failed to setup WireGuard", zap.Error(err))
		writeLog(fmt.Sprintf("Failed to setup WireGuard: %v", err))
		s.updateStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	}

	// 4. Setup Swarm
	writeLog("Setting up Swarm...")
	if err := s.setupSwarm(ctx, client, server); err != nil {
		logger.Log.Error("Failed to setup Swarm", zap.Error(err))
		writeLog(fmt.Sprintf("Failed to setup Swarm: %v", err))
		s.updateStatus(ctx, server.Id, entity.AgentStatusFailed)
		return
	}

	s.updateStatus(ctx, server.Id, entity.AgentStatusSuccess)
	writeLog("Node provisioning completed successfully")
	logger.Log.Info("Node provisioning completed successfully", zap.String("server_id", server.Id))

	// 5. Sync Mesh (Update all nodes with new peer)
	go s.SyncMesh(context.Background())
}

func (s *NodeService) setupSwarm(ctx context.Context, client *util.SSHClient, server *entity.Server) error {
	// Check if already in swarm
	if _, err := client.RunCommand("docker info --format '{{.Swarm.LocalNodeState}}' | grep active"); err == nil {
		return nil // Already active
	}

	return s.joinSwarm(ctx, client, server)
}

func (s *NodeService) initSwarm() error {
	// Check if local node is manager
	if err := exec.Command("docker", "node", "ls").Run(); err == nil {
		return nil
	}
	return exec.Command("docker", "swarm", "init").Run()
}

func (s *NodeService) getJoinToken() (string, error) {
	out, err := exec.Command("docker", "swarm", "join-token", "worker", "-q").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (s *NodeService) joinSwarm(ctx context.Context, client *util.SSHClient, server *entity.Server) error {
	token, err := s.getJoinToken()
	if err != nil {
		// Try init if token fails
		if err := s.initSwarm(); err != nil {
			return fmt.Errorf("failed to init swarm on controlplane: %w", err)
		}
		token, err = s.getJoinToken()
		if err != nil {
			return fmt.Errorf("failed to get join token: %w", err)
		}
	}

	managerIP := os.Getenv("SWARM_MANAGER_IP")
	if managerIP == "" {
		// Fallback to trying to detect or error
		return fmt.Errorf("SWARM_MANAGER_IP environment variable is not set")
	}

	// Join command
	// We use InternalIP (WG) for advertise and data-path to ensure traffic goes through VPN
	cmd := fmt.Sprintf("docker swarm join --token %s --advertise-addr %s --data-path-addr %s %s:2377",
		token, server.InternalIP, server.InternalIP, managerIP)

	if _, err := client.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to join swarm: %w", err)
	}

	// Set MTU to 1300 (WireGuard optimization)
	// We need to find the network interface created by Swarm (ingress/overlay) and set MTU.
	// But actually, we should set the default MTU for the docker daemon or network create.
	// For now, we rely on the `docker network create --opt com.docker.network.driver.mtu=1300` when creating networks.
	// But ingress network is created automatically.
	// We can update ingress network:
	// docker network update --opt com.docker.network.driver.mtu=1300 ingress (on manager)
	// But that's a one-time cluster setup.

	return nil
}

func (s *NodeService) setupWireGuard(ctx context.Context, client *util.SSHClient, server *entity.Server) error {
	// 1. Generate Keys
	privKey, err := client.RunCommand("wg genkey")
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}
	privKey = strings.TrimSpace(privKey)

	pubKey, err := client.RunCommand(fmt.Sprintf("echo '%s' | wg pubkey", privKey))
	if err != nil {
		return fmt.Errorf("failed to generate public key: %w", err)
	}
	pubKey = strings.TrimSpace(pubKey)

	// 2. Assign IP (Simple Strategy: 10.0.0.(count+2))
	// TODO: Implement robust IPAM
	servers, err := s.repo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to get servers for IP allocation: %w", err)
	}

	// Check if IP is already assigned to this server
	internalIP := server.InternalIP
	if internalIP == "" {
		// Very naive allocation, assumes no gaps and sequential IDs roughly map to count
		// In production, use a bitmap or DB table for IP pool.
		// We start from 10.0.0.2 (assuming .1 is gateway/controlplane if needed, or just start .1)
		// Let's use 10.0.0.(len+1)
		// We need to be careful about race conditions here.
		internalIP = fmt.Sprintf("10.0.0.%d", len(servers)+1)
	}

	// 3. Save to DB
	server.WireGuard.PrivateKey = privKey
	server.WireGuard.PublicKey = pubKey
	server.InternalIP = internalIP
	server.WireGuard.ListenPort = 51820

	if _, err := s.repo.Update(ctx, server); err != nil {
		return fmt.Errorf("failed to save WG keys: %w", err)
	}

	// 4. Create Config
	config := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s/24
ListenPort = %d
`, privKey, internalIP, server.WireGuard.ListenPort)

	// 5. Push Config
	if err := client.WriteFile("/etc/wireguard/wg0.conf", []byte(config), 0600); err != nil {
		return fmt.Errorf("failed to write wg0.conf: %w", err)
	}

	// 6. Enable
	if _, err := client.RunCommand("systemctl enable wg-quick@wg0 && systemctl restart wg-quick@wg0"); err != nil {
		return fmt.Errorf("failed to enable wireguard: %w", err)
	}

	return nil
}

func (s *NodeService) updateStatus(ctx context.Context, serverID string, status int) {
	server, err := s.repo.GetByID(ctx, serverID)
	if err != nil {
		logger.Log.Error("Failed to get server for status update", zap.Error(err))
		return
	}
	server.Agent.Status = status
	s.repo.Update(ctx, server)
}

func (s *NodeService) configureDockerDaemon(client *util.SSHClient) error {
	daemonConfig := `{
  "mtu": 1300
}`
	// Check if file exists to avoid overwriting custom config?
	// For now, we overwrite to ensure consistency.
	if err := client.WriteFile("/etc/docker/daemon.json", []byte(daemonConfig), 0644); err != nil {
		return fmt.Errorf("failed to write daemon.json: %w", err)
	}

	if _, err := client.RunCommand("systemctl restart docker"); err != nil {
		return fmt.Errorf("failed to restart docker: %w", err)
	}

	return nil
}

func (s *NodeService) SyncMesh(ctx context.Context) {
	// 1. Get all servers
	servers, err := s.repo.GetAll(ctx)
	if err != nil {
		logger.Log.Error("Failed to get servers for mesh sync", zap.Error(err))
		return
	}

	// 2. Filter valid peers (must have WG keys and IP)
	var validPeers []*entity.Server
	for _, srv := range servers {
		if srv.WireGuard.PublicKey != "" && srv.InternalIP != "" {
			validPeers = append(validPeers, srv)
		}
	}

	// 3. Sync each node
	// TODO: Use worker pool for parallelism
	for _, target := range validPeers {
		if err := s.syncNode(ctx, target, validPeers); err != nil {
			logger.Log.Error("Failed to sync node", zap.String("server_id", target.Id), zap.Error(err))
			// Continue to next node
		}
	}
}

func (s *NodeService) syncNode(ctx context.Context, target *entity.Server, peers []*entity.Server) error {
	client, err := util.NewSSHClient(target.IpAddress, target.Port, target.Credential.Username, target.Credential.Password, target.Credential.SSHKey)
	if err != nil {
		return err
	}
	defer client.Close()

	// Build Config
	var configBuilder strings.Builder

	// Interface
	configBuilder.WriteString(fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s/24
ListenPort = %d
`, target.WireGuard.PrivateKey, target.InternalIP, target.WireGuard.ListenPort))

	// Peers
	for _, peer := range peers {
		if peer.Id == target.Id {
			continue // Skip self
		}

		// Peer Block
		// Endpoint is the Public IP : ListenPort
		// AllowedIPs is the Internal IP / 32 (Host route)
		configBuilder.WriteString(fmt.Sprintf(`
[Peer]
# %s
PublicKey = %s
AllowedIPs = %s/32
Endpoint = %s:%d
PersistentKeepalive = 25
`, peer.Name, peer.WireGuard.PublicKey, peer.InternalIP, peer.IpAddress, peer.WireGuard.ListenPort))
	}

	// Write Config
	if err := client.WriteFile("/etc/wireguard/wg0.conf", []byte(configBuilder.String()), 0600); err != nil {
		return fmt.Errorf("failed to write wg0.conf: %w", err)
	}

	// Reload
	if _, err := client.RunCommand("systemctl restart wg-quick@wg0"); err != nil {
		return fmt.Errorf("failed to restart wireguard: %w", err)
	}

	return nil
}
