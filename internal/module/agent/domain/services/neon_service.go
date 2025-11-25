package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/zhinea/sylix/internal/common"
	"github.com/zhinea/sylix/internal/common/logger"
	"go.uber.org/zap"
)

const (
	DevComposeURL             = "https://raw.githubusercontent.com/zhinea/sylix/main/internal/module/agent/neon/docker-compose.yml"
	ReleaseComposeURLTemplate = "https://github.com/zhinea/sylix/releases/download/v%s/docker-compose.yml"
)

type NeonService struct {
	composeFile string
	httpClient  *http.Client
}

func NewNeonService(composeFile string) *NeonService {
	return &NeonService{
		composeFile: composeFile,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *NeonService) EnsureInfrastructure(ctx context.Context) error {
	logger.Log.Info("Ensuring Neon infrastructure is running...")

	// Always attempt to download/update the compose file
	logger.Log.Info("Downloading/Updating docker-compose file from GitHub", zap.String("path", s.composeFile))
	if err := s.downloadComposeFile(ctx); err != nil {
		// If download fails, check if we have a local copy to fall back on
		if _, statErr := os.Stat(s.composeFile); os.IsNotExist(statErr) {
			return fmt.Errorf("docker-compose file not found at %s and failed to download: %w", s.composeFile, err)
		}
		logger.Log.Warn("Failed to download docker-compose file, using local copy", zap.Error(err))
	}

	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", s.composeFile, "up", "-d")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Log.Error("Failed to start Neon infrastructure", zap.Error(err), zap.String("output", string(output)))
		return fmt.Errorf("failed to start neon infra: %w", err)
	}

	logger.Log.Info("Neon infrastructure started successfully")
	return nil
}

func (s *NeonService) downloadComposeFile(ctx context.Context) error {
	version := common.Version
	var url string

	if version == "0.0.0-dev" {
		url = DevComposeURL
	} else {
		url = fmt.Sprintf(ReleaseComposeURLTemplate, version)
	}

	logger.Log.Info("Downloading docker-compose file", zap.String("url", url))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: status %d", resp.StatusCode)
	}

	// Ensure directory exists
	dir := filepath.Dir(s.composeFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	out, err := os.Create(s.composeFile)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

type CreateTenantResponse struct {
	TenantID string `json:"id"` // The API returns just the ID string or an object? Usually object.
	// Adjust based on actual API response.
	// Neon API v1 returns { "id": "..." } usually?
	// Let's assume standard Neon API structure.
}

// Simplified response structure for parsing
type tenantResponse struct {
	ID string `json:"id"`
}

func (s *NeonService) CreateTenant(ctx context.Context) (string, error) {
	logger.Log.Info("Creating Neon Tenant")
	url := "http://localhost:9898/v1/tenant"
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Log.Error("Failed to execute CreateTenant request", zap.Error(err))
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Log.Error("CreateTenant returned error status", zap.Int("status", resp.StatusCode), zap.String("body", string(body)))
		return "", fmt.Errorf("failed to create tenant: status %d, body: %s", resp.StatusCode, string(body))
	}

	// The response might be just the ID string if using some versions, or a JSON object.
	// Let's try to parse as JSON string first (some older versions) or object.
	// Actually, standard Neon API returns a JSON object representing the tenant.
	// Example: {"id": "...", ...}

	// However, for local dev, sometimes it's different.
	// Let's assume it returns the ID in the body as a string if it's a simple dev tool,
	// but the real pageserver returns JSON.

	// Let's read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Try to unmarshal as string (if it's just "id")
	var id string
	if err := json.Unmarshal(body, &id); err == nil {
		return id, nil
	}

	// Try to unmarshal as object
	var result tenantResponse
	if err := json.Unmarshal(body, &result); err == nil && result.ID != "" {
		return result.ID, nil
	}

	// If unmarshal fails, maybe it's just the ID as raw string (unlikely for JSON API but possible)
	return string(bytes.TrimSpace(body)), nil
}

type timelineResponse struct {
	TenantID   string `json:"tenant_id"`
	TimelineID string `json:"timeline_id"`
}

func (s *NeonService) CreateTimeline(ctx context.Context, tenantID string, branchName string, parentTimelineID string) (string, error) {
	logger.Log.Info("Creating Neon Timeline", zap.String("tenant_id", tenantID), zap.String("branch", branchName))
	url := fmt.Sprintf("http://localhost:9898/v1/tenant/%s/timeline", tenantID)

	payload := map[string]interface{}{
		"new_timeline_id": nil,        // Auto-generate
		"timeline_id":     branchName, // This might be wrong. Usually timeline_id is UUID.
		// Neon API: POST /v1/tenant/:tenant_id/timeline
		// Body: { "new_timeline_id": "optional_uuid", "ancestor_timeline_id": "optional_uuid", "ancestor_lsn": "optional_lsn" }
	}

	if parentTimelineID != "" {
		payload["ancestor_timeline_id"] = parentTimelineID
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Log.Error("Failed to execute CreateTimeline request", zap.Error(err))
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Log.Error("CreateTimeline returned error status", zap.Int("status", resp.StatusCode), zap.String("body", string(body)))
		return "", fmt.Errorf("failed to create timeline: status %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result timelineResponse
	if err := json.Unmarshal(body, &result); err == nil && result.TimelineID != "" {
		return result.TimelineID, nil
	}

	// Fallback: try to parse just the ID if it returns just that (unlikely)
	var id string
	if err := json.Unmarshal(body, &id); err == nil {
		return id, nil
	}

	return "", fmt.Errorf("could not parse timeline response: %s", string(body))
}
