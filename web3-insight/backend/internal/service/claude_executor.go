package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"syscall"

	"github.com/google/uuid"
)

// ClaudeExecutor wraps Claude Code CLI execution
type ClaudeExecutor struct {
	sessionID string
}

// NewClaudeExecutor creates a new executor with a unique session ID
func NewClaudeExecutor() *ClaudeExecutor {
	return &ClaudeExecutor{
		sessionID: uuid.New().String(),
	}
}

// ClaudeResponse represents the JSON output from Claude Code
type ClaudeResponse struct {
	Result       string  `json:"result"`
	SessionID    string  `json:"session_id"`
	IsError      bool    `json:"is_error"`
	TotalCostUSD float64 `json:"total_cost_usd"`
}

// Execute runs a prompt through Claude Code CLI
func (e *ClaudeExecutor) Execute(ctx context.Context, prompt string) (*ClaudeResponse, error) {
	cmd := exec.CommandContext(ctx, "claude",
		"--print",
		"--session-id", e.sessionID,
		"--permission-mode", "bypassPermissions",
		"--output-format", "json",
		"--model", "sonnet",
		prompt,
	)

	// Set up process group for reliable killing
	// This ensures we can kill the entire process tree
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Ensure cleanup happens even if context is cancelled
	defer func() {
		if cmd.Process != nil {
			// Kill entire process group (negative PID = process group)
			// This ensures all child processes are terminated
			syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			cmd.Process.Kill() // Backup kill for main process
		}
	}()

	output, err := cmd.CombinedOutput()

	// Check if error was due to context timeout
	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("⏱️ TIMEOUT: Forcefully killed Claude process (session: %s) after timeout", e.sessionID)
		return nil, fmt.Errorf("command timeout: context deadline exceeded")
	}

	if err != nil {
		return nil, fmt.Errorf("claude execution failed: %w, output: %s", err, string(output))
	}

	var response ClaudeResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("parse response failed: %w, output: %s", err, string(output))
	}

	if response.IsError {
		return nil, fmt.Errorf("claude returned error: %s", response.Result)
	}

	return &response, nil
}

// GetSessionID returns the executor's session ID
func (e *ClaudeExecutor) GetSessionID() string {
	return e.sessionID
}
