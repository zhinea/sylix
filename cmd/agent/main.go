package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/zhinea/sylix/internal/common/config"
	"github.com/zhinea/sylix/internal/common/logger"
	agentPb "github.com/zhinea/sylix/internal/infra/proto/agent"
	grpcServices "github.com/zhinea/sylix/internal/module/agent/interface/grpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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

	agentService := grpcServices.NewAgentService()
	agentPb.RegisterAgentServer(grpcServer, agentService)

	logger.Log.Info("Agent started", zap.String("address", addr))
	if err := grpcServer.Serve(netListen); err != nil {
		logger.Log.Fatal("Failed to serve gRPC server", zap.String("address", addr), zap.Error(err))
	}
}
