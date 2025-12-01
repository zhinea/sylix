package app

import (
	"context"
	"time"

	"github.com/zhinea/sylix/internal/common/logger"
	"github.com/zhinea/sylix/internal/common/util"
	"github.com/zhinea/sylix/internal/module/controlplane/domain/repository"
	"github.com/zhinea/sylix/internal/module/controlplane/entity"
	"go.uber.org/zap"
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
	// TODO: Implement monitoring via SSH or Docker API
	// For now, we just check if we can SSH into the server
	client, err := util.NewSSHClient(server.IpAddress, server.Port, server.Credential.Username, server.Credential.Password, server.Credential.SSHKey)
	if err != nil {
		w.recordPingFailure(ctx, server, err.Error())
		return
	}
	defer client.Close()

	start := time.Now()
	// Simple command to check responsiveness
	if _, err := client.RunCommand("echo 'ping'"); err != nil {
		w.recordPingFailure(ctx, server, err.Error())
		return
	}
	duration := time.Since(start).Milliseconds()

	w.monitoringRepo.SavePing(ctx, &entity.ServerPing{
		ServerID:     server.Id,
		ResponseTime: duration,
		Status:       "OK",
	})

	// Check for high latency

}

func (w *MonitoringWorker) recordPingFailure(ctx context.Context, server *entity.Server, errorMsg string) {
	w.monitoringRepo.SavePing(ctx, &entity.ServerPing{
		ServerID:     server.Id,
		ResponseTime: 0,
		Status:       "ERROR",
		Error:        errorMsg,
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
