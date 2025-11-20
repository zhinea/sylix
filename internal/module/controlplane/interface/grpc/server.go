package grpc

import (
	"context"

	pbServer "github.com/zhinea/sylix/internal/infra/proto/server"
	"github.com/zhinea/sylix/internal/module/controlplane/entity"
	"github.com/zhinea/sylix/internal/module/controlplane/interface/grpc/validator"
	"gorm.io/gorm"
)

type ServerService struct {
	pbServer.UnimplementedServerServiceServer
	validator *validator.ServerValidator
	db        *gorm.DB
}

func NewServerService(db *gorm.DB) *ServerService {
	return &ServerService{
		validator: validator.NewServerValidator(),
		db:        db,
	}
}

func (s *ServerService) All(context.Context, *pbServer.Empty) (*pbServer.ServersResponse, error) {
	return &pbServer.ServersResponse{}, nil
}

func (s *ServerService) Get(context.Context, *pbServer.Id) (*pbServer.ServerResponse, error) {
	return &pbServer.ServerResponse{}, nil
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

	return &pbServer.ServerResponse{}, nil
}

func (s *ServerService) Update(context.Context, *pbServer.Server) (*pbServer.ServerResponse, error) {
	return &pbServer.ServerResponse{}, nil
}

func (s *ServerService) Delete(context.Context, *pbServer.Id) (*pbServer.MessageResponse, error) {
	return &pbServer.MessageResponse{}, nil
}

// Helper functions for conversion
func (s *ServerService) entityToProto(server *entity.Server) *pbServer.Server {
	// Implement conversion from entity to proto
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
	}
}

func (s *ServerService) protoToEntity(pb *pbServer.Server) *entity.Server {
	return &entity.Server{
		Name:      pb.Name,
		IpAddress: pb.IpAddress,
		Port:      int(pb.Port),
		Protocol:  pb.Protocol,
		Credential: entity.ServerCredential{
			Username: pb.Credential.Username,
			Password: pb.Credential.Password,
			SSHKey:   pb.Credential.SshKey,
		},
	}
}
