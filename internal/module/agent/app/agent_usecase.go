package app

import (
	"context"
	"os"
	"time"

	pbAgent "github.com/zhinea/sylix/internal/infra/proto/agent"
)

type AgentUseCase struct {
	startTime  time.Time
	configPath string
}

func NewAgentUseCase(configPath string) *AgentUseCase {
	return &AgentUseCase{
		startTime:  time.Now(),
		configPath: configPath,
	}
}

func (uc *AgentUseCase) GetStatus(ctx context.Context) (*pbAgent.GetStatusResponse, error) {
	return &pbAgent.GetStatusResponse{
		Status:  "RUNNING",
		Version: "0.1.0",
		Uptime:  int64(time.Since(uc.startTime).Seconds()),
	}, nil
}

func (uc *AgentUseCase) Heartbeat(ctx context.Context) (*pbAgent.HeartbeatResponse, error) {
	return &pbAgent.HeartbeatResponse{
		Acknowledged: true,
	}, nil
}

func (uc *AgentUseCase) Ping(ctx context.Context) (*pbAgent.PingResponse, error) {
	return &pbAgent.PingResponse{
		Timestamp: time.Now().Unix(),
		Status:    "OK",
	}, nil
}

func (uc *AgentUseCase) GetConfig(ctx context.Context) (*pbAgent.GetConfigResponse, error) {
	configContent, err := os.ReadFile(uc.configPath)
	if err != nil {
		return nil, err
	}
	return &pbAgent.GetConfigResponse{
		Config: string(configContent),
	}, nil
}
