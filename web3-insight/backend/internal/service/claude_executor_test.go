package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClaudeExecutor_Execute(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	executor := NewClaudeExecutor()
	assert.NotEmpty(t, executor.GetSessionID())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	prompt := `Return JSON: {"test": "success"}`

	response, err := executor.Execute(ctx, prompt)
	require.NoError(t, err, "Claude Code execution should succeed")
	require.NotNil(t, response)

	// Verify response structure
	assert.NotEmpty(t, response.Result, "Result should not be empty")
	assert.Equal(t, executor.GetSessionID(), response.SessionID, "Session ID should match")
	assert.False(t, response.IsError, "Should not be an error")

	// Try to parse the result as JSON to verify it contains our test data
	var resultData map[string]interface{}
	if err := json.Unmarshal([]byte(response.Result), &resultData); err == nil {
		// If result is valid JSON, check for our test data
		if testValue, ok := resultData["test"]; ok {
			assert.Equal(t, "success", testValue, "JSON should contain test field")
		}
	} else {
		// Result might be wrapped in text, that's okay for this test
		assert.Contains(t, response.Result, "success", "Result should contain 'success'")
	}

	t.Logf("Claude Code executed successfully. Cost: $%.4f", response.TotalCostUSD)
}

func TestClaudeExecutor_Execute_WithTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	executor := NewClaudeExecutor()

	// Very short timeout to test cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	prompt := "This should timeout before completing"

	_, err := executor.Execute(ctx, prompt)
	assert.Error(t, err, "Should fail with timeout")
}

func TestClaudeExecutor_SessionID_Unique(t *testing.T) {
	executor1 := NewClaudeExecutor()
	executor2 := NewClaudeExecutor()

	assert.NotEqual(t, executor1.GetSessionID(), executor2.GetSessionID(),
		"Each executor should have a unique session ID")
}
