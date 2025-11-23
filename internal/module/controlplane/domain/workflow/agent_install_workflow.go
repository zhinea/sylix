package workflow

import (
	"fmt"

	"github.com/zhinea/sylix/internal/common/workflow"
)

type AgentInstallParams struct {
	DownloadURL   string
	ServerCert    string
	ServerKey     string
	ConfigContent string
}

func NewAgentInstallWorkflow(params AgentInstallParams) workflow.Workflow {
	remoteBinaryPath := "/usr/local/bin/sylix-agent"

	// Service Content
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

	// Docker Install Script
	dockerInstallScript := `#!/bin/bash
set -e
export DEBIAN_FRONTEND=noninteractive

echo "Removing conflicting packages..."
sudo apt-get remove -y docker.io docker-compose docker-compose-v2 docker-doc podman-docker containerd runc || true

echo "Installing prerequisites..."
sudo apt-get update
sudo apt-get install -y ca-certificates curl

echo "Adding Docker GPG key..."
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc

echo "Adding Docker repository..."
sudo tee /etc/apt/sources.list.d/docker.sources <<EOF
Types: deb
URIs: https://download.docker.com/linux/ubuntu
Suites: $(. /etc/os-release && echo "${UBUNTU_CODENAME:-$VERSION_CODENAME}")
Components: stable
Signed-By: /etc/apt/keyrings/docker.asc
EOF

echo "Updating package index..."
sudo apt-get update

echo "Installing Docker..."
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

echo "Docker installation complete."
`

	return workflow.Workflow{
		Name: "Install Sylix Agent",
		Steps: []workflow.Step{
			{
				Name:        "Stop existing service",
				Action:      workflow.ActionCommand,
				Command:     "systemctl stop sylix-agent || true",
				IgnoreError: true,
			},
			{
				Name:    "Download agent binary",
				Action:  workflow.ActionCommand,
				Command: fmt.Sprintf("if command -v curl >/dev/null 2>&1; then curl -L -f -o %s %s; elif command -v wget >/dev/null 2>&1; then wget -O %s %s; else echo 'Error: neither curl nor wget found'; exit 1; fi", remoteBinaryPath, params.DownloadURL, remoteBinaryPath, params.DownloadURL),
			},
			{
				Name:    "Make agent executable",
				Action:  workflow.ActionCommand,
				Command: "chmod +x " + remoteBinaryPath,
			},
			{
				Name:    "Create certs directory",
				Action:  workflow.ActionCommand,
				Command: "mkdir -p /etc/sylix-agent/certs",
			},
			{
				Name:     "Write server certificate",
				Action:   workflow.ActionWriteFile,
				DestPath: "/etc/sylix-agent/certs/server.crt",
				Content:  params.ServerCert,
			},
			{
				Name:     "Write server key",
				Action:   workflow.ActionWriteFile,
				DestPath: "/etc/sylix-agent/certs/server.key",
				Content:  params.ServerKey,
			},
			{
				Name:    "Create config directory",
				Action:  workflow.ActionCommand,
				Command: "mkdir -p /etc/sylix-agent",
			},
			{
				Name:     "Write configuration file",
				Action:   workflow.ActionWriteFile,
				DestPath: "/etc/sylix-agent/config.yaml",
				Content:  params.ConfigContent,
			},
			{
				Name:     "Write systemd service file",
				Action:   workflow.ActionWriteFile,
				DestPath: "/etc/systemd/system/sylix-agent.service",
				Content:  serviceContent,
			},
			{
				Name:    "Reload systemd daemon",
				Action:  workflow.ActionCommand,
				Command: "systemctl daemon-reload",
			},
			{
				Name:    "Enable sylix-agent service",
				Action:  workflow.ActionCommand,
				Command: "systemctl enable sylix-agent",
			},
			{
				Name:    "Restart sylix-agent service",
				Action:  workflow.ActionCommand,
				Command: "systemctl restart sylix-agent",
			},
			// Docker Installation Steps
			{
				Name:      "Write Docker install script",
				Action:    workflow.ActionWriteFile,
				DestPath:  "/tmp/install_docker.sh",
				Content:   dockerInstallScript,
				Condition: "! command -v docker",
			},
			{
				Name:      "Run Docker install script",
				Action:    workflow.ActionCommand,
				Command:   "chmod +x /tmp/install_docker.sh && sudo /tmp/install_docker.sh",
				Condition: "! command -v docker",
			},
			{
				Name:      "Cleanup Docker install script",
				Action:    workflow.ActionCommand,
				Command:   "rm /tmp/install_docker.sh",
				Condition: "! command -v docker",
			},
		},
	}
}
