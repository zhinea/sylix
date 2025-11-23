package workflow

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/zhinea/sylix/internal/common/util"
)

// ActionType defines what a step does
type ActionType string

const (
	ActionCommand   ActionType = "command"    // Run a shell command
	ActionWriteFile ActionType = "write_file" // Write content to a remote file
	ActionCopyFile  ActionType = "copy_file"  // Copy a local file to remote
)

type Step struct {
	Name   string
	Action ActionType

	// Command specific
	Command string // The shell command to run

	// File specific
	Content    string // Content for WriteFile
	SourcePath string // Local path for CopyFile
	DestPath   string // Remote path

	// Control flow
	// Condition is a shell command. If it returns exit code 0, the step runs.
	// If empty, the step always runs.
	Condition   string
	IgnoreError bool // If true, continues even if the step fails
}

type Workflow struct {
	Name  string
	Steps []Step
}

type Engine struct {
	client    *util.SSHClient
	logWriter io.Writer
	logFn     func(string)
}

func NewEngine(client *util.SSHClient, logWriter io.Writer, logFn func(string)) *Engine {
	return &Engine{
		client:    client,
		logWriter: logWriter,
		logFn:     logFn,
	}
}

func (e *Engine) Run(ctx context.Context, wf Workflow) error {
	e.logFn(fmt.Sprintf("Starting workflow: %s", wf.Name))

	for i, step := range wf.Steps {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		e.logFn(fmt.Sprintf("[Step %d/%d] %s", i+1, len(wf.Steps), step.Name))

		// Check condition if present
		if step.Condition != "" {
			// We use RunCommand (not stream) to check condition silently,
			// but maybe we want to log it?
			// For now, let's run it silently to check exit code.
			// If we want to log the check, we can use RunCommandStream but discard output if successful?
			// The requirement says "if Docker is not installed...".
			// Usually condition checks are silent unless they fail or we want debug info.
			// Let's run it. If it fails (non-zero exit code), we skip the step.
			// Wait, the requirement says: "if Docker is not installed, the installation should continue"
			// So if Condition is "check if docker missing", and it returns true (0), we run the step.

			// We can use RunCommand. If it returns error, it means exit code != 0.
			_, err := e.client.RunCommand(step.Condition)
			if err != nil {
				e.logFn(fmt.Sprintf("Skipping step '%s' (condition not met)", step.Name))
				continue
			}
		}

		if err := e.executeStep(step); err != nil {
			if step.IgnoreError {
				e.logFn(fmt.Sprintf("Step '%s' failed but marked to ignore error: %v", step.Name, err))
				continue
			}
			return fmt.Errorf("step '%s' failed: %w", step.Name, err)
		}
	}

	e.logFn(fmt.Sprintf("Workflow '%s' completed successfully.", wf.Name))
	return nil
}

func (e *Engine) executeStep(step Step) error {
	switch step.Action {
	case ActionCommand:
		return e.client.RunCommandStream(step.Command, e.logWriter, e.logWriter)

	case ActionWriteFile:
		// Create a temporary file locally
		tmpFile := fmt.Sprintf("workflow_tmp_%d", time.Now().UnixNano())
		if err := os.WriteFile(tmpFile, []byte(step.Content), 0644); err != nil {
			return fmt.Errorf("failed to write temp file: %w", err)
		}
		defer os.Remove(tmpFile)

		if err := e.client.CopyFile(tmpFile, step.DestPath); err != nil {
			return fmt.Errorf("failed to copy file to remote: %w", err)
		}
		return nil

	case ActionCopyFile:
		if err := e.client.CopyFile(step.SourcePath, step.DestPath); err != nil {
			return fmt.Errorf("failed to copy file to remote: %w", err)
		}
		return nil

	default:
		return fmt.Errorf("unknown action type: %s", step.Action)
	}
}
