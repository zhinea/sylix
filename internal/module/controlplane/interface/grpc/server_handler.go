package grpc

import (
	"context"

	"github.com/zhinea/sylix/internal/module/controlplane/app"
	"github.com/zhinea/sylix/internal/module/controlplane/entity"
	"github.com/zhinea/sylix/internal/module/controlplane/interface/grpc/validator"

	pbServer "github.com/zhinea/sylix/internal/infra/proto/server"
)

type ServerService struct {
	pbServer.UnimplementedServerServiceServer
	validator *validator.ServerValidator
	useCase   *app.ServerUseCase
}

func NewServerService(useCase *app.ServerUseCase) *ServerService {
	return &ServerService{
		validator: validator.NewServerValidator(),
		useCase:   useCase,
	}
}

func (s *ServerService) All(ctx context.Context, _ *pbServer.Empty) (*pbServer.ServersResponse, error) {
	servers, err := s.useCase.GetAll(ctx)
	if err != nil {
		errStr := err.Error()
		return &pbServer.ServersResponse{
			Status: pbServer.StatusCode_INTERNAL_ERROR,
			Error:  &errStr,
		}, nil
	}

	var pbServers []*pbServer.Server
	for _, server := range servers {
		pbServers = append(pbServers, s.entityToProto(server))
	}

	return &pbServer.ServersResponse{
		Status:  pbServer.StatusCode_OK,
		Servers: pbServers,
	}, nil
}

func (s *ServerService) Get(ctx context.Context, id *pbServer.Id) (*pbServer.ServerResponse, error) {
	server, err := s.useCase.Get(ctx, id.Id)
	if err != nil {
		errStr := err.Error()
		return &pbServer.ServerResponse{
			Status: pbServer.StatusCode_NOT_FOUND,
			Error:  &errStr,
		}, nil
	}

	return &pbServer.ServerResponse{
		Status: pbServer.StatusCode_OK,
		Server: s.entityToProto(server),
	}, nil
}

func (s *ServerService) Create(ctx context.Context, pb *pbServer.Server) (*pbServer.ServerResponse, error) {
	// Convert protobuf server to entity server
	entityServer := s.protoToEntity(pb)

	if err := s.validator.Validate(entityServer); err != nil {
		return &pbServer.ServerResponse{
			Status: pbServer.StatusCode_VALIDATION_FAILED,
			Server: &pbServer.Server{},
			Errors: err,
		}, nil
	}

	createdServer, err := s.useCase.Create(ctx, entityServer)
	if err != nil {
		errStr := err.Error()
		return &pbServer.ServerResponse{
			Status: pbServer.StatusCode_INTERNAL_ERROR,
			Error:  &errStr,
		}, nil
	}

	return &pbServer.ServerResponse{
		Status: pbServer.StatusCode_CREATED,
		Server: s.entityToProto(createdServer),
	}, nil
}

func (s *ServerService) Update(ctx context.Context, pb *pbServer.Server) (*pbServer.ServerResponse, error) {
	entityServer := s.protoToEntity(pb)

	updatedServer, err := s.useCase.Update(ctx, entityServer)
	if err != nil {
		errStr := err.Error()
		return &pbServer.ServerResponse{
			Status: pbServer.StatusCode_INTERNAL_ERROR,
			Error:  &errStr,
		}, nil
	}

	return &pbServer.ServerResponse{
		Status: pbServer.StatusCode_OK,
		Server: s.entityToProto(updatedServer),
	}, nil
}

func (s *ServerService) Delete(ctx context.Context, id *pbServer.Id) (*pbServer.MessageResponse, error) {
	if err := s.useCase.Delete(ctx, id.Id); err != nil {
		return &pbServer.MessageResponse{
			Status:  pbServer.StatusCode_INTERNAL_ERROR,
			Message: err.Error(),
		}, nil
	}

	return &pbServer.MessageResponse{
		Status:  pbServer.StatusCode_OK,
		Message: "Server deleted successfully",
	}, nil
}

func (s *ServerService) InstallAgent(ctx context.Context, id *pbServer.Id) (*pbServer.MessageResponse, error) {
	if err := s.useCase.InstallAgent(ctx, id.Id); err != nil {
		return &pbServer.MessageResponse{
			Status:  pbServer.StatusCode_INTERNAL_ERROR,
			Message: err.Error(),
		}, nil
	}

	return &pbServer.MessageResponse{
		Status:  pbServer.StatusCode_OK,
		Message: "Agent installed successfully",
	}, nil
}

// Helper functions for conversion
var errStr = "Internal Server Error" // Placeholder for error string pointer

func (s *ServerService) entityToProto(server *entity.Server) *pbServer.Server {
	return &pbServer.Server{
		Id:        server.Id,
		Name:      server.Name,
		IpAddress: server.IpAddress,
		Port:      int32(server.Port),
		Protocol:  server.Protocol,
		Credential: &pbServer.ServerCredential{
			Username: server.Credential.Username,
			Password: server.Credential.Password,
			SshKey:   server.Credential.SSHKey,
		},
		Status:      pbServer.StatusServer(server.Status),
		AgentStatus: pbServer.AgentStatusServer(server.AgentStatus),
		AgentLogs:   server.AgentLogs,
	}
}

func (s *ServerService) protoToEntity(pb *pbServer.Server) *entity.Server {
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
		Status:      int(pb.Status),
		AgentStatus: int(pb.AgentStatus),
		AgentLogs:   pb.AgentLogs,
	}
	server.Id = pb.Id
	return server
}
