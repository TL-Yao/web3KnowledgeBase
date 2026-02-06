package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKBScheduler_StartStop(t *testing.T) {
	// Create a mock orchestrator (nil is fine for this test)
	scheduler := NewKBScheduler(nil)

	// Initially not running
	assert.False(t, scheduler.IsRunning(), "Scheduler should not be running initially")
	assert.Nil(t, scheduler.GetNextRunTime(), "Next run time should be nil when not running")

	// Start scheduler
	err := scheduler.Start()
	require.NoError(t, err, "Should start scheduler successfully")
	assert.True(t, scheduler.IsRunning(), "Scheduler should be running after start")

	// Verify next run time is set
	nextRun := scheduler.GetNextRunTime()
	assert.NotNil(t, nextRun, "Next run time should be set when running")
	assert.True(t, nextRun.After(time.Now()), "Next run should be in the future")

	// Starting again should be idempotent
	err = scheduler.Start()
	require.NoError(t, err, "Starting again should not error")
	assert.True(t, scheduler.IsRunning(), "Scheduler should still be running")

	// Stop scheduler
	scheduler.Stop()
	assert.False(t, scheduler.IsRunning(), "Scheduler should not be running after stop")

	// Stopping again should be safe
	scheduler.Stop()
	assert.False(t, scheduler.IsRunning(), "Scheduler should still be stopped")
}

func TestKBScheduler_GetSchedulerStatus(t *testing.T) {
	scheduler := NewKBScheduler(nil)

	// Status when not running
	status := scheduler.GetSchedulerStatus()
	assert.False(t, status.IsRunning)
	assert.Equal(t, "Every 4 hours (0 */4 * * *)", status.Schedule)
	assert.Nil(t, status.NextRun)

	// Status when running
	err := scheduler.Start()
	require.NoError(t, err)
	defer scheduler.Stop()

	status = scheduler.GetSchedulerStatus()
	assert.True(t, status.IsRunning)
	assert.Equal(t, "Every 4 hours (0 */4 * * *)", status.Schedule)
	assert.NotNil(t, status.NextRun)
}

// Note: Full integration test for scheduled execution would require:
// - Real orchestrator with working dependencies
// - Waiting for scheduled time (expensive)
// - Mock time.Now() or use very short intervals
//
// This is better suited for manual testing or E2E test suite.
