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
	// Maps to ProvisionNode
	server, err := s.useCase.Get(ctx, id.Id)
	if err != nil {
		return &pbControlPlane.MessageResponse{
			Status:  pbControlPlane.StatusCode_NOT_FOUND,
			Message: err.Error(),
		}, nil
	}

	s.useCase.ProvisionNode(ctx, server)

	return &pbControlPlane.MessageResponse{
		Status:  pbControlPlane.StatusCode_OK,
		Message: "Node provisioning started",
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
		Status: pbControlPlane.StatusServer(server.Status),
		Agent: &pbControlPlane.ServerAgent{
			Port:   int32(server.Agent.Port),
			Status: pbControlPlane.AgentStatusServer(server.Agent.Status),
			Logs:   server.Agent.Logs,
		},
	}
}

func (s *ServerService) protoToEntity(pb *pbControlPlane.Server) *entity.Server {
	server := &entity.Server{
		Name:      pb.Name,
		IpAddress: pb.IpAddress,
		Port:      int(pb.Port),
		Protocol:  pb.Protocol,
		Credential: entity.ServerCredential{
			Username: pb.Credential.Username,
			Password: pb.Credential.Password,
			SSHKey:   pb.Credential.SshKey,
		},
		Status: int(pb.Status),
	}
	if pb.Agent != nil {
		server.Agent = entity.ServerAgent{
			Port:   int(pb.Agent.Port),
			Status: int(pb.Agent.Status),
			Logs:   pb.Agent.Logs,
		}
	}
	server.Id = pb.Id
	return server
}
