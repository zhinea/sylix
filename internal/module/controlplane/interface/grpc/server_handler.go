package grpc

import (
	"context"
	"time"

	"github.com/zhinea/sylix/internal/module/controlplane/app"
	"github.com/zhinea/sylix/internal/module/controlplane/entity"
	"github.com/zhinea/sylix/internal/module/controlplane/interface/grpc/validator"

	pbCommon "github.com/zhinea/sylix/internal/infra/proto/common"
	pbControlPlane "github.com/zhinea/sylix/internal/infra/proto/controlplane"
)

type ServerService struct {
	pbControlPlane.UnimplementedServerServiceServer
	validator *validator.ServerValidator
	useCase   *app.ServerUseCase
}

func NewServerService(useCase *app.ServerUseCase) *ServerService {
	return &ServerService{
		validator: validator.NewServerValidator(),
		useCase:   useCase,
	}
}

func (s *ServerService) All(ctx context.Context, _ *pbCommon.Empty) (*pbControlPlane.ServersResponse, error) {
	servers, err := s.useCase.GetAll(ctx)
	if err != nil {
		errStr := err.Error()
		return &pbControlPlane.ServersResponse{
			Status: pbControlPlane.StatusCode_INTERNAL_ERROR,
			Error:  &errStr,
		}, nil
	}

	var pbControlPlanes []*pbControlPlane.Server
	for _, server := range servers {
		pbControlPlanes = append(pbControlPlanes, s.entityToProto(server))
	}

	return &pbControlPlane.ServersResponse{
		Status:  pbControlPlane.StatusCode_OK,
		Servers: pbControlPlanes,
	}, nil
}

func (s *ServerService) Get(ctx context.Context, id *pbControlPlane.Id) (*pbControlPlane.ServerResponse, error) {
	server, err := s.useCase.Get(ctx, id.Id)
	if err != nil {
		errStr := err.Error()
		return &pbControlPlane.ServerResponse{
			Status: pbControlPlane.StatusCode_NOT_FOUND,
			Error:  &errStr,
		}, nil
	}

	return &pbControlPlane.ServerResponse{
		Status: pbControlPlane.StatusCode_OK,
		Server: s.entityToProto(server),
	}, nil
}

func (s *ServerService) Create(ctx context.Context, pb *pbControlPlane.Server) (*pbControlPlane.ServerResponse, error) {
	// Convert protobuf server to entity server
	entityServer := s.protoToEntity(pb)

	if err := s.validator.Validate(entityServer); err != nil {
		return &pbControlPlane.ServerResponse{
			Status: pbControlPlane.StatusCode_VALIDATION_FAILED,
			Server: &pbControlPlane.Server{},
			Errors: err,
		}, nil
	}

	createdServer, err := s.useCase.Create(ctx, entityServer)
	if err != nil {
		errStr := err.Error()
		return &pbControlPlane.ServerResponse{
			Status: pbControlPlane.StatusCode_INTERNAL_ERROR,
			Error:  &errStr,
		}, nil
	}

	return &pbControlPlane.ServerResponse{
		Status: pbControlPlane.StatusCode_CREATED,
		Server: s.entityToProto(createdServer),
	}, nil
}

func (s *ServerService) Update(ctx context.Context, pb *pbControlPlane.Server) (*pbControlPlane.ServerResponse, error) {
	entityServer := s.protoToEntity(pb)

	updatedServer, err := s.useCase.Update(ctx, entityServer)
	if err != nil {
		errStr := err.Error()
		return &pbControlPlane.ServerResponse{
			Status: pbControlPlane.StatusCode_INTERNAL_ERROR,
			Error:  &errStr,
		}, nil
	}

	return &pbControlPlane.ServerResponse{
		Status: pbControlPlane.StatusCode_OK,
		Server: s.entityToProto(updatedServer),
	}, nil
}

func (s *ServerService) GetStats(ctx context.Context, req *pbControlPlane.GetStatsRequest) (*pbControlPlane.GetStatsResponse, error) {
	stats, err := s.useCase.GetStats(ctx, req.ServerId)
	if err != nil {
		return nil, err
	}

	var pbStats []*pbControlPlane.ServerStat
	for _, stat := range stats {
		pbStats = append(pbStats, &pbControlPlane.ServerStat{
			Id:                  stat.Id,
			ServerId:            stat.ServerID,
			AverageResponseTime: stat.AverageResponseTime,
			MinResponseTime:     stat.MinResponseTime,
			MaxResponseTime:     stat.MaxResponseTime,
			PingCount:           stat.PingCount,
			SuccessRate:         stat.SuccessRate,
			Timestamp:           stat.Timestamp.Format(time.RFC3339),
		})
	}

	return &pbControlPlane.GetStatsResponse{
		Stats: pbStats,
	}, nil
}

func (s *ServerService) GetAccidents(ctx context.Context, req *pbControlPlane.GetAccidentsRequest) (*pbControlPlane.GetAccidentsResponse, error) {
	var startDate, endDate *time.Time
	if req.StartDate != "" {
		t, err := time.Parse(time.RFC3339, req.StartDate)
		if err == nil {
			startDate = &t
		}
	}
	if req.EndDate != "" {
		t, err := time.Parse(time.RFC3339, req.EndDate)
		if err == nil {
			endDate = &t
		}
	}

	// Handle resolved filter if needed (not in proto yet, assuming all for now or adding logic)
	// The proto has `bool resolved = 4;` but bool defaults to false, so we need a way to know if it was set.
	// For now, let's assume we fetch all if not specified, or we can add a flag.
	// Since proto3 doesn't have optional scalars easily without `optional` keyword, let's assume we pass a pointer if we want to filter.
	// But here we receive a value. Let's ignore resolved filter for a moment or assume the user always sends it.
	// Better approach: Use a separate field `filter_resolved` boolean to indicate if we should filter by `resolved`.
	// For this implementation, I'll pass nil for resolved to fetch all.

	accidents, total, err := s.useCase.GetAccidents(ctx, req.ServerId, startDate, endDate, nil, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, err
	}

	var pbAccidents []*pbControlPlane.ServerAccident
	for _, accident := range accidents {
		pbAccidents = append(pbAccidents, &pbControlPlane.ServerAccident{
			Id:           accident.Id,
			ServerId:     accident.ServerID,
			ResponseTime: accident.ResponseTime,
			Error:        accident.Error,
			Details:      accident.Details,
			Resolved:     accident.Resolved,
			CreatedAt:    accident.CreatedAt.Format(time.RFC3339),
		})
	}

	return &pbControlPlane.GetAccidentsResponse{
		Accidents:  pbAccidents,
		TotalCount: total,
	}, nil
}

func (s *ServerService) GetRealtimeStats(ctx context.Context, req *pbControlPlane.GetRealtimeStatsRequest) (*pbControlPlane.GetRealtimeStatsResponse, error) {
	pings, err := s.useCase.GetRealtimeStats(ctx, req.ServerId, int(req.Limit))
	if err != nil {
		return nil, err
	}

	var pbPings []*pbControlPlane.ServerPing
	for _, ping := range pings {
		pbPings = append(pbPings, &pbControlPlane.ServerPing{
			Id:           ping.Id,
			ServerId:     ping.ServerID,
			ResponseTime: ping.ResponseTime,
			Status:       ping.Status,
			Error:        ping.Error,
			CreatedAt:    ping.CreatedAt.Format(time.RFC3339),
		})
	}

	return &pbControlPlane.GetRealtimeStatsResponse{
		Pings: pbPings,
	}, nil
}

func (s *ServerService) ConfigureAgent(ctx context.Context, req *pbControlPlane.ConfigureAgentRequest) (*pbControlPlane.MessageResponse, error) {
	if err := s.useCase.ConfigureAgent(ctx, req.ServerId, req.Config); err != nil {
		return &pbControlPlane.MessageResponse{
			Status:  pbControlPlane.StatusCode_INTERNAL_ERROR,
			Message: err.Error(),
		}, nil
	}
	return &pbControlPlane.MessageResponse{
		Status:  pbControlPlane.StatusCode_OK,
		Message: "Agent configured successfully",
	}, nil
}

func (s *ServerService) UpdateAgentPort(ctx context.Context, req *pbControlPlane.UpdateAgentPortRequest) (*pbControlPlane.MessageResponse, error) {
	if err := s.useCase.UpdateAgentPort(ctx, req.ServerId, int(req.Port)); err != nil {
		return &pbControlPlane.MessageResponse{
			Status:  pbControlPlane.StatusCode_INTERNAL_ERROR,
			Message: err.Error(),
		}, nil
	}
	return &pbControlPlane.MessageResponse{
		Status:  pbControlPlane.StatusCode_OK,
		Message: "Agent port updated successfully",
	}, nil
}

func (s *ServerService) UpdateServerTimeZone(ctx context.Context, req *pbControlPlane.UpdateServerTimeZoneRequest) (*pbControlPlane.MessageResponse, error) {
	if err := s.useCase.UpdateServerTimeZone(ctx, req.ServerId, req.Timezone); err != nil {
		return &pbControlPlane.MessageResponse{
			Status:  pbControlPlane.StatusCode_INTERNAL_ERROR,
			Message: err.Error(),
		}, nil
	}
	return &pbControlPlane.MessageResponse{
		Status:  pbControlPlane.StatusCode_OK,
		Message: "Server timezone updated successfully",
	}, nil
}

func (s *ServerService) Delete(ctx context.Context, id *pbControlPlane.Id) (*pbControlPlane.MessageResponse, error) {
	if err := s.useCase.Delete(ctx, id.Id); err != nil {
		return &pbControlPlane.MessageResponse{
			Status:  pbControlPlane.StatusCode_INTERNAL_ERROR,
			Message: err.Error(),
		}, nil
	}

	return &pbControlPlane.MessageResponse{
		Status:  pbControlPlane.StatusCode_OK,
		Message: "Server deleted successfully",
	}, nil
}

func (s *ServerService) InstallAgent(ctx context.Context, id *pbControlPlane.Id) (*pbControlPlane.MessageResponse, error) {
	if err := s.useCase.InstallAgent(ctx, id.Id); err != nil {
		return &pbControlPlane.MessageResponse{
			Status:  pbControlPlane.StatusCode_INTERNAL_ERROR,
			Message: err.Error(),
		}, nil
	}

	return &pbControlPlane.MessageResponse{
		Status:  pbControlPlane.StatusCode_OK,
		Message: "Agent installed successfully",
	}, nil
}

func (s *ServerService) GetAgentConfig(ctx context.Context, id *pbControlPlane.Id) (*pbControlPlane.GetAgentConfigResponse, error) {
	config, timezone, err := s.useCase.GetAgentConfig(ctx, id.Id)
	if err != nil {
		return nil, err
	}

	return &pbControlPlane.GetAgentConfigResponse{
		Config:   config,
		Timezone: timezone,
	}, nil
}

func (s *ServerService) DeleteAccident(ctx context.Context, req *pbControlPlane.Id) (*pbControlPlane.MessageResponse, error) {
	err := s.useCase.DeleteAccident(ctx, req.Id)
	if err != nil {
		return &pbControlPlane.MessageResponse{
			Status:  pbControlPlane.StatusCode_INTERNAL_ERROR,
			Message: err.Error(),
		}, nil
	}
	return &pbControlPlane.MessageResponse{
		Status:  pbControlPlane.StatusCode_OK,
		Message: "Accident deleted successfully",
	}, nil
}

func (s *ServerService) BatchDeleteAccidents(ctx context.Context, req *pbControlPlane.BatchDeleteAccidentsRequest) (*pbControlPlane.MessageResponse, error) {
	err := s.useCase.BatchDeleteAccidents(ctx, req.Ids)
	if err != nil {
		return &pbControlPlane.MessageResponse{
			Status:  pbControlPlane.StatusCode_INTERNAL_ERROR,
			Message: err.Error(),
		}, nil
	}
	return &pbControlPlane.MessageResponse{
		Status:  pbControlPlane.StatusCode_OK,
		Message: "Accidents deleted successfully",
	}, nil
}

// Helper functions for conversion
var errStr = "Internal Server Error" // Placeholder for error string pointer

func (s *ServerService) entityToProto(server *entity.Server) *pbControlPlane.Server {
	return &pbControlPlane.Server{
		Id:        server.Id,
		Name:      server.Name,
		IpAddress: server.IpAddress,
		Port:      int32(server.Port),
		Protocol:  server.Protocol,
		Credential: &pbControlPlane.ServerCredential{
			Username: server.Credential.Username,
			Password: server.Credential.Password,
			SshKey:   server.Credential.SSHKey,
		},
		Status:      pbControlPlane.StatusServer(server.Status),
		AgentStatus: pbControlPlane.AgentStatusServer(server.AgentStatus),
		AgentLogs:   server.AgentLogs,
		AgentPort:   int32(server.AgentPort),
	}
}

func (s *ServerService) protoToEntity(pb *pbControlPlane.Server) *entity.Server {
	server := &entity.Server{
		Name:      pb.Name,
		IpAddress: pb.IpAddress,
		Port:      int(pb.Port),
		AgentPort: int(pb.AgentPort),
		Protocol:  pb.Protocol,
		Credential: entity.ServerCredential{
			Username: pb.Credential.Username,
			Password: pb.Credential.Password,
			SSHKey:   pb.Credential.SshKey,
		},
		Status:      int(pb.Status),
		AgentStatus: int(pb.AgentStatus),
		AgentLogs:   pb.AgentLogs,
	}
	server.Id = pb.Id
	return server
}
