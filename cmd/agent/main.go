package main

import (
	"net"

	"github.com/zhinea/sylix/internal/common/logger"
	agentPb "github.com/zhinea/sylix/internal/infra/proto/agent"
	grpcServices "github.com/zhinea/sylix/internal/module/agent/interface/grpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	logger.Init(logger.Config{
		Filename:   "sylix-agent.log",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	})
	defer logger.Log.Sync()

	port := ":8083"

	netListen, err := net.Listen("tcp", port)
	if err != nil {
		logger.Log.Fatal("Failed to listen", zap.String("port", port), zap.Error(err))
	}

	grpcServer := grpc.NewServer()

	agentService := grpcServices.NewAgentService()
	agentPb.RegisterAgentServer(grpcServer, agentService)

	logger.Log.Info("Agent started", zap.String("port", port))
	if err := grpcServer.Serve(netListen); err != nil {
		logger.Log.Fatal("Failed to serve gRPC server", zap.String("port", port), zap.Error(err))
	}
}
