package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/zhinea/sylix/internal/common/config"
	"github.com/zhinea/sylix/internal/common/logger"
	agentPb "github.com/zhinea/sylix/internal/infra/proto/agent"
	"github.com/zhinea/sylix/internal/module/agent/app"
	"github.com/zhinea/sylix/internal/module/agent/domain/services"
	grpcServices "github.com/zhinea/sylix/internal/module/agent/interface/grpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	configPath := flag.String("config", "/etc/sylix-agent/config.yaml", "Path to configuration file")
	flag.Parse()

	cfg, err := config.LoadAgentConfig(*configPath)
	if err != nil {
		// Fallback to basic logging if config load fails
		logger.Init(logger.Config{Filename: "sylix-agent-startup.log"})
		logger.Log.Fatal("Failed to load configuration", zap.Error(err))
	}

	logger.Init(logger.Config{
		Filename:   cfg.Log.Filename,
		MaxSize:    cfg.Log.MaxSize,
		MaxBackups: cfg.Log.MaxBackups,
		MaxAge:     cfg.Log.MaxAge,
		Compress:   cfg.Log.Compress,
		Level:      cfg.Log.Level,
	})
	defer logger.Log.Sync()

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	netListen, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Log.Fatal("Failed to listen", zap.String("address", addr), zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	if cfg.Security.CertFile != "" && cfg.Security.KeyFile != "" {
		creds, err := credentials.NewServerTLSFromFile(cfg.Security.CertFile, cfg.Security.KeyFile)
		if err != nil {
			logger.Log.Fatal("Failed to load TLS keys", zap.Error(err))
		}
		grpcServer = grpc.NewServer(grpc.Creds(creds))
	}

	dockerService, err := services.NewDockerService()
	if err != nil {
		logger.Log.Fatal("Failed to create Docker service", zap.Error(err))
	}

	// Initialize Neon Service
	// Assuming running from project root for dev
	neonComposeFile := "internal/module/agent/neon/docker-compose.yml"
	neonService := services.NewNeonService(neonComposeFile)

	agentUseCase := app.NewAgentUseCase(*configPath, dockerService, neonService)
	agentHandler := grpcServices.NewAgentHandler(agentUseCase)
	agentPb.RegisterAgentServer(grpcServer, agentHandler)

	logger.Log.Info("Agent started", zap.String("address", addr))
	if err := grpcServer.Serve(netListen); err != nil {
		logger.Log.Fatal("Failed to serve gRPC server", zap.String("address", addr), zap.Error(err))
	}
}
