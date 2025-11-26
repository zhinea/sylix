package grpc

import (
	"context"

	pbAgent "github.com/zhinea/sylix/internal/infra/proto/agent"
	"github.com/zhinea/sylix/internal/module/agent/app"
)

type AgentHandler struct {
	pbAgent.UnimplementedAgentServer
	useCase *app.AgentUseCase
}

func NewAgentHandler(useCase *app.AgentUseCase) *AgentHandler {
	return &AgentHandler{
		useCase: useCase,
	}
}

func (h *AgentHandler) GetStatus(ctx context.Context, req *pbAgent.GetStatusRequest) (*pbAgent.GetStatusResponse, error) {
	return h.useCase.GetStatus(ctx)
}

func (h *AgentHandler) Heartbeat(ctx context.Context, req *pbAgent.HeartbeatRequest) (*pbAgent.HeartbeatResponse, error) {
	return h.useCase.Heartbeat(ctx)
}

func (h *AgentHandler) Ping(ctx context.Context, req *pbAgent.PingRequest) (*pbAgent.PingResponse, error) {
	return h.useCase.Ping(ctx)
}

func (h *AgentHandler) GetConfig(ctx context.Context, req *pbAgent.GetConfigRequest) (*pbAgent.GetConfigResponse, error) {
	return h.useCase.GetConfig(ctx)
}
