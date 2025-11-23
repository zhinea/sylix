package app

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"time"

	"github.com/zhinea/sylix/internal/common/logger"
	pbAgent "github.com/zhinea/sylix/internal/infra/proto/agent"
	"github.com/zhinea/sylix/internal/module/controlplane/domain/repository"
	"github.com/zhinea/sylix/internal/module/controlplane/entity"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type MonitoringWorker struct {
	serverRepo     repository.ServerRepository
	monitoringRepo repository.MonitoringRepository
}

func NewMonitoringWorker(serverRepo repository.ServerRepository, monitoringRepo repository.MonitoringRepository) *MonitoringWorker {
	return &MonitoringWorker{
		serverRepo:     serverRepo,
		monitoringRepo: monitoringRepo,
	}
}

func (w *MonitoringWorker) Start() {
	go w.runPingLoop()
	go w.runStatsLoop()
}

func (w *MonitoringWorker) runPingLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		w.pingAllServers()
	}
}

func (w *MonitoringWorker) runStatsLoop() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		w.calculateStats()
		w.cleanupOldPings()
	}
}

func (w *MonitoringWorker) pingAllServers() {
	ctx := context.Background()
	servers, err := w.serverRepo.GetAll(ctx)
	if err != nil {
		logger.Log.Error("Failed to get servers for monitoring", zap.Error(err))
		return
	}

	for _, server := range servers {
		// Only ping connected servers
		if server.Status == entity.ServerStatusDisconnected {
			continue
		}
		go w.pingServer(ctx, server)
	}
}

func (w *MonitoringWorker) pingServer(ctx context.Context, server *entity.Server) {
	port := server.AgentPort
	if port == 0 {
		port = 8083
	}
	target := fmt.Sprintf("%s:%d", server.IpAddress, port)

	var opts []grpc.DialOption
	if server.Credential.CaCert != "" {
		cp := x509.NewCertPool()
		if !cp.AppendCertsFromPEM([]byte(server.Credential.CaCert)) {
			w.recordPingFailure(ctx, server, "failed to append CA cert")
			return
		}
		// We use InsecureSkipVerify: true here ONLY for the hostname check if we want to allow IP-based certs easily without strict hostname matching issues,
		// BUT since we generated the cert with the IP as SAN, we should be fine.
		// However, grpc-go's TLS credentials expect the server name to match.
		// If we don't pass a server name override, it uses the target address.
		config := &tls.Config{
			RootCAs:            cp,
			InsecureSkipVerify: true,
		}
		creds := credentials.NewTLS(config)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		w.recordPingFailure(ctx, server, err.Error())
		return
	}
	defer conn.Close()

	client := pbAgent.NewAgentClient(conn)

	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err = client.Ping(ctx, &pbAgent.PingRequest{Timestamp: start.Unix()})
	duration := time.Since(start).Milliseconds()

	if err != nil {
		w.recordPingFailure(ctx, server, err.Error())
		return
	}

	w.monitoringRepo.SavePing(ctx, &entity.ServerPing{
		ServerID:     server.Id,
		ResponseTime: duration,
		Status:       "OK",
	})

	// Check for high latency
	if duration > 500 {
		w.monitoringRepo.SaveAccident(ctx, &entity.ServerAccident{
			ServerID:     server.Id,
			ResponseTime: duration,
			Error:        "High Latency",
			Details:      fmt.Sprintf("Response time %dms > 500ms", duration),
			Resolved:     false,
		})
	}
}

func (w *MonitoringWorker) recordPingFailure(ctx context.Context, server *entity.Server, errorMsg string) {
	w.monitoringRepo.SavePing(ctx, &entity.ServerPing{
		ServerID:     server.Id,
		ResponseTime: 0,
		Status:       "ERROR",
		Error:        errorMsg,
	})

	w.monitoringRepo.SaveAccident(ctx, &entity.ServerAccident{
		ServerID:     server.Id,
		ResponseTime: 0,
		Error:        "Connection Failed",
		Details:      errorMsg,
		Resolved:     false,
	})
}

func (w *MonitoringWorker) calculateStats() {
	ctx := context.Background()
	servers, err := w.serverRepo.GetAll(ctx)
	if err != nil {
		logger.Log.Error("Failed to get servers for stats", zap.Error(err))
		return
	}

	for _, server := range servers {
		// Get pings for the last 15 minutes
		since := time.Now().Add(-15 * time.Minute)
		pings, err := w.monitoringRepo.GetPingsByServerID(ctx, server.Id, since)
		if err != nil {
			continue
		}

		if len(pings) == 0 {
			continue
		}

		var totalTime int64
		var minTime int64 = -1
		var maxTime int64
		var successCount int64

		for _, p := range pings {
			if p.Status == "OK" {
				totalTime += p.ResponseTime
				if minTime == -1 || p.ResponseTime < minTime {
					minTime = p.ResponseTime
				}
				if p.ResponseTime > maxTime {
					maxTime = p.ResponseTime
				}
				successCount++
			}
		}

		if successCount == 0 {
			continue
		}

		avgTime := float64(totalTime) / float64(successCount)
		successRate := float64(successCount) / float64(len(pings)) * 100

		w.monitoringRepo.SaveStat(ctx, &entity.ServerStat{
			ServerID:            server.Id,
			AverageResponseTime: avgTime,
			MinResponseTime:     minTime,
			MaxResponseTime:     maxTime,
			PingCount:           int64(len(pings)),
			SuccessRate:         successRate,
			Timestamp:           time.Now(),
		})
	}
}

func (w *MonitoringWorker) cleanupOldPings() {
	ctx := context.Background()
	// Keep pings for 24 hours
	before := time.Now().Add(-24 * time.Hour)
	if err := w.monitoringRepo.DeleteOldPings(ctx, before); err != nil {
		logger.Log.Error("Failed to cleanup old pings", zap.Error(err))
	}
}
