package grpc

import (
	"context"
	"os"
	"strings"
	"time"

	pbAgent "github.com/zhinea/sylix/internal/infra/proto/agent"
)

type AgentService struct {
	pbAgent.UnimplementedAgentServer
	startTime  time.Time
	configPath string
}

func NewAgentService(configPath string) *AgentService {
	return &AgentService{
		startTime:  time.Now(),
		configPath: configPath,
	}
}

func (s *AgentService) GetStatus(ctx context.Context, req *pbAgent.GetStatusRequest) (*pbAgent.GetStatusResponse, error) {
	return &pbAgent.GetStatusResponse{
		Status:  "RUNNING",
		Version: "0.1.0",
		Uptime:  int64(time.Since(s.startTime).Seconds()),
	}, nil
}

func (s *AgentService) Heartbeat(ctx context.Context, req *pbAgent.HeartbeatRequest) (*pbAgent.HeartbeatResponse, error) {
	// Log heartbeat or update status
	return &pbAgent.HeartbeatResponse{
		Acknowledged: true,
	}, nil
}

func (s *AgentService) Ping(ctx context.Context, req *pbAgent.PingRequest) (*pbAgent.PingResponse, error) {
	return &pbAgent.PingResponse{
		Timestamp: time.Now().Unix(),
		Status:    "OK",
	}, nil
}

func (s *AgentService) GetConfig(ctx context.Context, req *pbAgent.GetConfigRequest) (*pbAgent.GetConfigResponse, error) {
	configContent, err := os.ReadFile(s.configPath)
	if err != nil {
		return nil, err
	}

	timezone := "UTC"
	if tzBytes, err := os.ReadFile("/etc/timezone"); err == nil {
		timezone = strings.TrimSpace(string(tzBytes))
	} else if link, err := os.Readlink("/etc/localtime"); err == nil {
		parts := strings.Split(link, "zoneinfo/")
		if len(parts) > 1 {
			timezone = parts[1]
		}
	}

	return &pbAgent.GetConfigResponse{
		Config:   string(configContent),
		Timezone: timezone,
	}, nil
}
