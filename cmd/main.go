package main

import (
	"log"
	"net"

	database "github.com/zhinea/sylix/internal/infra/db"
	serverPb "github.com/zhinea/sylix/internal/infra/proto/server"
	"github.com/zhinea/sylix/internal/module/controlplane/app"
	"github.com/zhinea/sylix/internal/module/controlplane/domain/repository"
	grpcServices "github.com/zhinea/sylix/internal/module/controlplane/interface/grpc"
	"google.golang.org/grpc"
)

func main() {
	db, err := database.NewDB()

	if err != nil {
		panic(err)
	}

	database.AutoMigrate(db)

	port := ":8082"

	netListen, err := net.Listen("tcp", port)

	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()

	// Initialize dependencies
	serverRepo := repository.NewServerRepository(db)
	serverUseCase := app.NewServerUseCase(serverRepo)
	serverService := grpcServices.NewServerService(serverUseCase)

	serverPb.RegisterServerServiceServer(grpcServer, serverService)

	log.Printf("Server started at: %v", port)
	if err := grpcServer.Serve(netListen); err != nil {
		log.Fatalf("Failed to serve gRPC server over port %s: %v", port, err)
	}
}
