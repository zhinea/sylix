package main

import (
	"log"
	"net/http"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"github.com/zhinea/sylix/internal/common/logger"
	database "github.com/zhinea/sylix/internal/infra/db"
	pbControlPlane "github.com/zhinea/sylix/internal/infra/proto/controlplane"
	"github.com/zhinea/sylix/internal/module/controlplane/app"
	"github.com/zhinea/sylix/internal/module/controlplane/domain/repository"
	"github.com/zhinea/sylix/internal/module/controlplane/domain/services"
	grpcServices "github.com/zhinea/sylix/internal/module/controlplane/interface/grpc"
	"google.golang.org/grpc"
)

func main() {
	_ = godotenv.Load() // Load .env file if it exists

	logger.Init(logger.Config{
		Filename:   "logs/app/file.log",
		MaxSize:    10, // MB
		MaxBackups: 3,
		MaxAge:     7, // days
		Compress:   true,
	})
	defer logger.Log.Sync()

	db, err := database.NewDB()

	if err != nil {
		panic(err)
	}

	database.AutoMigrate(db)

	port := ":8082"

	grpcServer := grpc.NewServer()

	// Initialize dependencies
	serverRepo := repository.NewServerRepository(db)
	monitoringRepo := repository.NewMonitoringRepository(db)
	backupRepo := repository.NewBackupStorageRepository(db)

	monitoringService := services.NewMonitoringService(monitoringRepo)
	agentService := services.NewAgentService(serverRepo)
	backupService := services.NewBackupService(backupRepo, serverRepo, agentService)

	serverUseCase := app.NewServerUseCase(serverRepo, monitoringService, agentService)
	serverService := grpcServices.NewServerService(serverUseCase)
	backupStorageService := grpcServices.NewBackupStorageService(backupService)

	logsUseCase := app.NewLogsUseCase()
	logsService := grpcServices.NewLogsService(logsUseCase)

	// Monitoring
	monitoringWorker := app.NewMonitoringWorker(serverRepo, monitoringRepo)
	monitoringWorker.Start()

	pbControlPlane.RegisterServerServiceServer(grpcServer, serverService)
	pbControlPlane.RegisterLogsServiceServer(grpcServer, logsService)
	pbControlPlane.RegisterBackupStorageServiceServer(grpcServer, backupStorageService)

	// Wrap gRPC server for gRPC-Web support
	wrappedGrpc := grpcweb.WrapServer(grpcServer,
		grpcweb.WithOriginFunc(func(origin string) bool {
			return true // Allow all origins for development
		}),
	)

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// Create HTTP handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if wrappedGrpc.IsGrpcWebRequest(r) {
			wrappedGrpc.ServeHTTP(w, r)
			return
		}
		// Fallback to standard gRPC server (if using HTTP/2) or other handlers
		// Note: serving standard gRPC over HTTP/1.1 port usually requires cmux or h2c
		// For now, we prioritize gRPC-Web for the dashboard.
		wrappedGrpc.ServeHTTP(w, r)
	})

	httpServer := &http.Server{
		Addr:    port,
		Handler: c.Handler(handler),
	}

	log.Printf("Server started at: %v", port)
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
