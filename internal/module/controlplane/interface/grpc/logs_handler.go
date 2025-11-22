package grpc

import (
	"context"
	"os"

	pbLogs "github.com/zhinea/sylix/internal/infra/proto/logs"
	"github.com/zhinea/sylix/internal/module/controlplane/app"
)

type LogsService struct {
	pbLogs.UnimplementedLogsServiceServer
	useCase *app.LogsUseCase
}

func NewLogsService(useCase *app.LogsUseCase) *LogsService {
	return &LogsService{
		useCase: useCase,
	}
}

func (s *LogsService) GetServerLogs(ctx context.Context, req *pbLogs.GetServerLogsRequest) (*pbLogs.GetServerLogsResponse, error) {
	entries, err := s.useCase.GetServerLogs(ctx, req.ServerId)
	if err != nil {
		return nil, err
	}

	var files []*pbLogs.LogFile
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, &pbLogs.LogFile{
			Name:         entry.Name(),
			Size:         info.Size(),
			LastModified: info.ModTime().String(),
		})
	}

	return &pbLogs.GetServerLogsResponse{
		Files: files,
	}, nil
}

func (s *LogsService) ReadServerLog(ctx context.Context, req *pbLogs.ReadServerLogRequest) (*pbLogs.ReadServerLogResponse, error) {
	lines, totalLines, totalPages, err := s.useCase.ReadServerLog(ctx, req.ServerId, req.Filename, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, err
	}

	return &pbLogs.ReadServerLogResponse{
		Lines:       lines,
		TotalLines:  int32(totalLines),
		CurrentPage: req.Page,
		TotalPages:  int32(totalPages),
	}, nil
}

func (s *LogsService) GetSystemLogs(ctx context.Context, _ *pbLogs.Empty) (*pbLogs.GetSystemLogsResponse, error) {
	logFiles, err := s.useCase.GetSystemLogs(ctx)
	if err != nil {
		return nil, err
	}

	var files []*pbLogs.LogFile
	for _, path := range logFiles {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		files = append(files, &pbLogs.LogFile{
			Name:         path,
			Size:         info.Size(),
			LastModified: info.ModTime().String(),
		})
	}

	return &pbLogs.GetSystemLogsResponse{
		Files: files,
	}, nil
}

func (s *LogsService) ReadSystemLog(ctx context.Context, req *pbLogs.ReadSystemLogRequest) (*pbLogs.ReadSystemLogResponse, error) {
	lines, totalLines, totalPages, err := s.useCase.ReadSystemLog(ctx, req.Filename, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, err
	}

	return &pbLogs.ReadSystemLogResponse{
		Lines:       lines,
		TotalLines:  int32(totalLines),
		CurrentPage: req.Page,
		TotalPages:  int32(totalPages),
	}, nil
}
