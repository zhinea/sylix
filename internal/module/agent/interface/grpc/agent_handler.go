package grpc

import (
	"context"
	"time"

	pbAgent "github.com/zhinea/sylix/internal/infra/proto/agent"
)

type AgentService struct {
	pbAgent.UnimplementedAgentServer
	startTime time.Time
}

func NewAgentService() *AgentService {
	return &AgentService{
		startTime: time.Now(),
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
