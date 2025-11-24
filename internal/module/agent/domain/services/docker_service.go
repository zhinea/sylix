package services

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/zhinea/sylix/internal/common/logger"
	"go.uber.org/zap"
)

type DockerService struct {
	cli *client.Client
}

func NewDockerService() (*DockerService, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithVersion("1.44"))
	if err != nil {
		return nil, err
	}
	return &DockerService{cli: cli}, nil
}

func (s *DockerService) PullImage(ctx context.Context, imageName string) error {
	logger.Log.Debug("DockerService: Pulling image", zap.String("image", imageName))
	reader, err := s.cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		logger.Log.Error("DockerService: Failed to pull image", zap.Error(err))
		return err
	}
	defer reader.Close()
	io.Copy(os.Stdout, reader) // Or log it
	return nil
}

func (s *DockerService) CreateAndStartContainer(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, containerName string) (string, error) {
	// Check if container exists
	_, err := s.cli.ContainerInspect(ctx, containerName)
	if err == nil {
		// Container exists, maybe start it if stopped?
		// For now, let's assume we want to fail or return existing ID if it matches.
		// But simpler to just try create and handle error if needed.
		logger.Log.Warn("DockerService: Container already exists", zap.String("container_name", containerName))
	}

	resp, err := s.cli.ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
	if err != nil {
		logger.Log.Error("DockerService: Failed to create container", zap.Error(err), zap.String("container_name", containerName))
		return "", err
	}

	if err := s.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		logger.Log.Error("DockerService: Failed to start container", zap.Error(err), zap.String("container_id", resp.ID))
		return "", err
	}

	logger.Log.Debug("DockerService: Container started", zap.String("container_id", resp.ID))
	return resp.ID, nil
}

func (s *DockerService) GetContainerPort(ctx context.Context, containerID string, internalPort string) (string, error) {
	inspect, err := s.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		logger.Log.Error("DockerService: Failed to inspect container", zap.Error(err), zap.String("container_id", containerID))
		return "", err
	}

	bindings := inspect.NetworkSettings.Ports[nat.Port(internalPort)]
	if len(bindings) > 0 {
		hostPort := bindings[0].HostPort
		logger.Log.Debug("DockerService: Port bound", zap.String("internal", internalPort), zap.String("host", hostPort))
		return hostPort, nil
	}
	logger.Log.Error("DockerService: Port not bound", zap.String("internal", internalPort))
	return "", fmt.Errorf("port not bound")
}
