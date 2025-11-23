package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type LogsUseCase struct{}

func NewLogsUseCase() *LogsUseCase {
	return &LogsUseCase{}
}

func (uc *LogsUseCase) GetServerLogs(ctx context.Context, serverID string) ([]os.DirEntry, error) {
	logDir := fmt.Sprintf("logs/servers/%s", serverID)
	entries, err := os.ReadDir(logDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []os.DirEntry{}, nil
		}
		return nil, err
	}
	return entries, nil
}

func (uc *LogsUseCase) ReadServerLog(ctx context.Context, serverID, filename string, page, pageSize int) ([]string, int, int, int, error) {
	logPath := fmt.Sprintf("logs/servers/%s/%s", serverID, filename)
	// Security check: ensure filename doesn't contain ".."
	if filepath.Base(filename) != filename {
		return nil, 0, 0, 0, fmt.Errorf("invalid filename")
	}

	return uc.readLogFile(logPath, page, pageSize)
}

func (uc *LogsUseCase) GetSystemLogs(ctx context.Context) ([]string, error) {
	var logs []string
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		return logs, nil
	}
	err := filepath.Walk("logs", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			logs = append(logs, path)
		}
		return nil
	})
	return logs, err
}

func (uc *LogsUseCase) ReadSystemLog(ctx context.Context, filename string, page, pageSize int) ([]string, int, int, int, error) {
	cleanPath := filepath.Clean(filename)

	if !filepath.HasPrefix(cleanPath, "logs"+string(os.PathSeparator)) && cleanPath != "logs" {
		return nil, 0, 0, 0, fmt.Errorf("invalid log path: must start with logs/")
	}

	if filepath.IsAbs(cleanPath) {
		return nil, 0, 0, 0, fmt.Errorf("absolute paths not allowed")
	}

	// Verify it is inside logs directory
	absLogs, _ := filepath.Abs("logs")
	absPath, _ := filepath.Abs(cleanPath)
	if !filepath.HasPrefix(absPath, absLogs) {
		return nil, 0, 0, 0, fmt.Errorf("access denied")
	}

	return uc.readLogFile(cleanPath, page, pageSize)
}

func (uc *LogsUseCase) readLogFile(path string, page, pageSize int) ([]string, int, int, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	totalLines := len(lines)
	totalPages := (totalLines + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if start > totalLines {
		start = totalLines
	}
	if end > totalLines {
		end = totalLines
	}

	return lines[start:end], totalLines, totalPages, page, nil
}
