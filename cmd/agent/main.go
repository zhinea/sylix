package main

import (
	"log"
	"net"

	agentPb "github.com/zhinea/sylix/internal/infra/proto/agent"
	grpcServices "github.com/zhinea/sylix/internal/module/agent/interface/grpc"
	"google.golang.org/grpc"
)

func main() {
	port := ":8083"

	netListen, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()

	agentService := grpcServices.NewAgentService()
	agentPb.RegisterAgentServer(grpcServer, agentService)

	log.Printf("Agent started at: %v", port)
	if err := grpcServer.Serve(netListen); err != nil {
		log.Fatalf("Failed to serve gRPC server over port %s: %v", port, err)
	}
}
