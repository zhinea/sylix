package workflow

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
