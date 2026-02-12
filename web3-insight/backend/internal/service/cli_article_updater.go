package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/user/web3-insight/internal/llm"
	"github.com/user/web3-insight/internal/model"
)

const cliUpdateTimeout = 5 * time.Minute

// CLIArticleUpdater generates article updates via Claude Code CLI (subscription auth, $0 cost)
type CLIArticleUpdater struct{}

// NewCLIArticleUpdater creates a new CLI-based article updater
func NewCLIArticleUpdater() *CLIArticleUpdater {
	return &CLIArticleUpdater{}
}

// GenerateUpdate generates a targeted article update using Claude Code CLI
func (u *CLIArticleUpdater) GenerateUpdate(
	ctx context.Context,
	article *model.Article,
	conversationHistory []llm.Message,
) (*UpdateResult, error) {
	prompt := buildUpdatePrompt(article, conversationHistory)

	executor := NewClaudeExecutorWithOptions(ClaudeExecutorOptions{
		SystemPrompt: articleUpdateSystemPrompt,
		StripAPIKey:  true,
		Model:        "sonnet",
	})

	log.Printf("CLI article updater: generating update for article %s (session: %s)", article.ID, executor.GetSessionID())

	genCtx, cancel := context.WithTimeout(ctx, cliUpdateTimeout)
	defer cancel()

	response, err := executor.Execute(genCtx, prompt)
	if err != nil {
		return nil, fmt.Errorf("CLI execution failed: %w", err)
	}

	log.Printf("CLI article updater: completed for article %s, cost: $%.4f", article.ID, response.TotalCostUSD)

	updatedContent, changeSummary, err := parseUpdateResponse(response.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CLI response: %w", err)
	}

	return &UpdateResult{
		UpdatedContent: updatedContent,
		ChangeSummary:  changeSummary,
		Model:          "claude-cli (subscription)",
	}, nil
}

// CLIAvailable checks if the Claude CLI is available on the system
func CLIAvailable() bool {
	executor := NewClaudeExecutorWithOptions(ClaudeExecutorOptions{
		StripAPIKey: true,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := executor.Execute(ctx, "respond with exactly: OK")
	if err != nil {
		return false
	}
	return strings.Contains(response.Result, "OK")
}
