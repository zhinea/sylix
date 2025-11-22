package grpc

import (
	"context"
	"os"

	pbCommon "github.com/zhinea/sylix/internal/infra/proto/common"
	pbControlPlane "github.com/zhinea/sylix/internal/infra/proto/controlplane"
	"github.com/zhinea/sylix/internal/module/controlplane/app"
)

type LogsService struct {
	pbControlPlane.UnimplementedLogsServiceServer
	useCase *app.LogsUseCase
}

func NewLogsService(useCase *app.LogsUseCase) *LogsService {
	return &LogsService{
		useCase: useCase,
	}
}

func (s *LogsService) GetServerLogs(ctx context.Context, req *pbControlPlane.GetServerLogsRequest) (*pbControlPlane.GetServerLogsResponse, error) {
	entries, err := s.useCase.GetServerLogs(ctx, req.ServerId)
	if err != nil {
		return nil, err
	}

	var files []*pbControlPlane.LogFile
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, &pbControlPlane.LogFile{
			Name:         entry.Name(),
			Size:         info.Size(),
			LastModified: info.ModTime().String(),
		})
	}

	return &pbControlPlane.GetServerLogsResponse{
		Files: files,
	}, nil
}

func (s *LogsService) ReadServerLog(ctx context.Context, req *pbControlPlane.ReadServerLogRequest) (*pbControlPlane.ReadServerLogResponse, error) {
	lines, totalLines, totalPages, err := s.useCase.ReadServerLog(ctx, req.ServerId, req.Filename, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, err
	}

	return &pbControlPlane.ReadServerLogResponse{
		Lines:       lines,
		TotalLines:  int32(totalLines),
		CurrentPage: req.Page,
		TotalPages:  int32(totalPages),
	}, nil
}

func (s *LogsService) GetSystemLogs(ctx context.Context, _ *pbCommon.Empty) (*pbControlPlane.GetSystemLogsResponse, error) {
	logFiles, err := s.useCase.GetSystemLogs(ctx)
	if err != nil {
		return nil, err
	}

	var files []*pbControlPlane.LogFile
	for _, path := range logFiles {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		files = append(files, &pbControlPlane.LogFile{
			Name:         path,
			Size:         info.Size(),
			LastModified: info.ModTime().String(),
		})
	}

	return &pbControlPlane.GetSystemLogsResponse{
		Files: files,
	}, nil
}

func (s *LogsService) ReadSystemLog(ctx context.Context, req *pbControlPlane.ReadSystemLogRequest) (*pbControlPlane.ReadSystemLogResponse, error) {
	lines, totalLines, totalPages, err := s.useCase.ReadSystemLog(ctx, req.Filename, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, err
	}

	return &pbControlPlane.ReadSystemLogResponse{
		Lines:       lines,
		TotalLines:  int32(totalLines),
		CurrentPage: req.Page,
		TotalPages:  int32(totalPages),
	}, nil
}
