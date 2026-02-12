package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/google/uuid"
)

// ClaudeExecutorOptions configures ClaudeExecutor behavior
type ClaudeExecutorOptions struct {
	SystemPrompt   string // If set, passed via --system-prompt flag
	StripAPIKey    bool   // If true, removes ANTHROPIC_API_KEY from env (forces subscription auth)
	Model          string // Model override (default: "sonnet")
}

// ClaudeExecutor wraps Claude Code CLI execution
type ClaudeExecutor struct {
	sessionID string
	opts      ClaudeExecutorOptions
}

// NewClaudeExecutor creates a new executor with a unique session ID
func NewClaudeExecutor() *ClaudeExecutor {
	return &ClaudeExecutor{
		sessionID: uuid.New().String(),
	}
}

// NewClaudeExecutorWithOptions creates a new executor with custom options
func NewClaudeExecutorWithOptions(opts ClaudeExecutorOptions) *ClaudeExecutor {
	return &ClaudeExecutor{
		sessionID: uuid.New().String(),
		opts:      opts,
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
	model := e.opts.Model
	if model == "" {
		model = "sonnet"
	}

	args := []string{
		"--print",
		"--session-id", e.sessionID,
		"--permission-mode", "bypassPermissions",
		"--output-format", "json",
		"--model", model,
	}

	if e.opts.SystemPrompt != "" {
		args = append(args, "--system-prompt", e.opts.SystemPrompt)
	}

	args = append(args, prompt)

	cmd := exec.CommandContext(ctx, "claude", args...)

	// Set up process group for reliable killing
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Filter environment if StripAPIKey is set (forces subscription auth)
	if e.opts.StripAPIKey {
		cmd.Env = filterEnv(os.Environ(), "ANTHROPIC_API_KEY")
	}

	// Ensure cleanup happens even if context is cancelled
	defer func() {
		if cmd.Process != nil {
			syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			cmd.Process.Kill()
		}
	}()

	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("TIMEOUT: Forcefully killed Claude process (session: %s) after timeout", e.sessionID)
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

// filterEnv returns a copy of env with entries matching the given key prefix removed
func filterEnv(env []string, key string) []string {
	prefix := key + "="
	filtered := make([]string, 0, len(env))
	for _, e := range env {
		if !strings.HasPrefix(e, prefix) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}
