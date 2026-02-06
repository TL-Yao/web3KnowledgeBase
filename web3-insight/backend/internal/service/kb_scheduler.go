package service

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/user/web3-insight/internal/model"
)

// KBScheduler manages scheduled knowledge base updates
type KBScheduler struct {
	orchestrator *KBUpdateOrchestrator
	cron         *cron.Cron
	isRunning    bool
}

// NewKBScheduler creates a new scheduler instance
func NewKBScheduler(orchestrator *KBUpdateOrchestrator) *KBScheduler {
	return &KBScheduler{
		orchestrator: orchestrator,
		cron:         cron.New(),
		isRunning:    false,
	}
}

// Start begins the scheduled updates (every 4 hours)
func (s *KBScheduler) Start() error {
	if s.isRunning {
		log.Println("Scheduler is already running")
		return nil
	}

	// Schedule: every 4 hours
	// Cron expression: "0 */4 * * *" = at minute 0 of every 4th hour
	_, err := s.cron.AddFunc("0 */4 * * *", func() {
		log.Println("Scheduled KB update triggered")
		s.runScheduledUpdate()
	})

	if err != nil {
		return err
	}

	s.cron.Start()
	s.isRunning = true
	log.Println("KB update scheduler started (runs every 4 hours)")

	return nil
}

// Stop stops the scheduler
func (s *KBScheduler) Stop() {
	if !s.isRunning {
		log.Println("Scheduler is not running")
		return
	}

	s.cron.Stop()
	s.isRunning = false
	log.Println("KB update scheduler stopped")
}

// IsRunning returns whether the scheduler is currently running
func (s *KBScheduler) IsRunning() bool {
	return s.isRunning
}

// runScheduledUpdate executes a scheduled update
func (s *KBScheduler) runScheduledUpdate() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()

	job, err := s.orchestrator.RunUpdate(ctx, "scheduled")
	if err != nil {
		log.Printf("Scheduled KB update failed: %v", err)
		return
	}

	log.Printf("Scheduled KB update completed: job_id=%s, articles=%d/%d",
		job.ID, job.ArticlesGenerated, job.KeywordsGenerated)
}

// GetNextRunTime returns the time of the next scheduled run
func (s *KBScheduler) GetNextRunTime() *time.Time {
	if !s.isRunning {
		return nil
	}

	entries := s.cron.Entries()
	if len(entries) == 0 {
		return nil
	}

	nextRun := entries[0].Next
	return &nextRun
}

// GetSchedulerStatus returns current scheduler status
func (s *KBScheduler) GetSchedulerStatus() SchedulerStatus {
	status := SchedulerStatus{
		IsRunning: s.isRunning,
		Schedule:  "Every 4 hours (0 */4 * * *)",
	}

	if s.isRunning {
		nextRun := s.GetNextRunTime()
		if nextRun != nil {
			status.NextRun = nextRun
		}
	}

	return status
}

// SchedulerStatus represents the current state of the scheduler
type SchedulerStatus struct {
	IsRunning bool       `json:"is_running"`
	Schedule  string     `json:"schedule"`
	NextRun   *time.Time `json:"next_run,omitempty"`
}

// TriggerManual triggers an immediate manual update (doesn't affect schedule)
func (s *KBScheduler) TriggerManual(ctx context.Context) (*model.KBUpdateJob, error) {
	log.Println("Manual KB update triggered (via scheduler)")
	return s.orchestrator.RunUpdate(ctx, "manual")
}
