package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
)

const (
	DefaultKeywordBatchSize = 3 // Number of keywords to process per update (reduced for testing/cost savings)
)

type KBUpdateOrchestrator struct {
	keywordPool  *KeywordPoolService
	articleGen   *ArticleGeneratorService
	keywordRepo  *repository.KeywordRepository
	jobRepo      *repository.KBUpdateJobRepository
}

func NewKBUpdateOrchestrator(
	keywordPool *KeywordPoolService,
	articleGen *ArticleGeneratorService,
	keywordRepo *repository.KeywordRepository,
	jobRepo *repository.KBUpdateJobRepository,
) *KBUpdateOrchestrator {
	return &KBUpdateOrchestrator{
		keywordPool:  keywordPool,
		articleGen:   articleGen,
		keywordRepo:  keywordRepo,
		jobRepo:      jobRepo,
	}
}

// RunUpdate executes a complete knowledge base update cycle
func (ko *KBUpdateOrchestrator) RunUpdate(ctx context.Context, triggerType string) (*model.KBUpdateJob, error) {
	log.Printf("Starting knowledge base update (trigger: %s)", triggerType)

	// 0. Check for concurrent job execution (prevent multiple simultaneous updates)
	runningJobs, err := ko.jobRepo.FindByStatus("running")
	if err != nil {
		return nil, fmt.Errorf("failed to check running jobs: %w", err)
	}

	if len(runningJobs) > 0 {
		return nil, fmt.Errorf("update already in progress (job: %s), please wait for completion", runningJobs[0].ID)
	}

	// 1. Cleanup orphaned jobs (mark jobs running > 30 minutes as failed)
	cleanedCount, err := ko.jobRepo.CleanupOrphanedJobs(30 * time.Minute)
	if err != nil {
		log.Printf("Warning: Failed to cleanup orphaned jobs: %v", err)
	} else if cleanedCount > 0 {
		log.Printf("Cleaned up %d orphaned jobs", cleanedCount)
	}

	// 2. Create job record
	now := time.Now()
	job := &model.KBUpdateJob{
		Status:      "running",
		TriggerType: triggerType,
		StartedAt:   &now,
	}

	if err := ko.jobRepo.Create(job); err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	log.Printf("Created job %s", job.ID)

	// 3. Check and refill keyword pool
	log.Println("Checking keyword pool...")
	if err := ko.keywordPool.RefillPoolIfNeeded(ctx); err != nil {
		ko.jobRepo.UpdateStatus(job.ID, "failed", fmt.Sprintf("keyword pool refill failed: %v", err))
		return nil, fmt.Errorf("failed to refill keyword pool: %w", err)
	}

	// 4. Get pending keywords
	log.Printf("Fetching %d pending keywords...", DefaultKeywordBatchSize)
	keywords, err := ko.keywordRepo.GetPendingKeywords(DefaultKeywordBatchSize)
	if err != nil {
		ko.jobRepo.UpdateStatus(job.ID, "failed", fmt.Sprintf("failed to fetch keywords: %v", err))
		return nil, fmt.Errorf("failed to get pending keywords: %w", err)
	}

	if len(keywords) < DefaultKeywordBatchSize {
		errMsg := fmt.Sprintf("insufficient keywords: got %d, need %d", len(keywords), DefaultKeywordBatchSize)
		ko.jobRepo.UpdateStatus(job.ID, "failed", errMsg)
		return nil, fmt.Errorf("insufficient keywords: got %d, need %d", len(keywords), DefaultKeywordBatchSize)
	}

	job.KeywordsGenerated = len(keywords)
	ko.jobRepo.Update(job)

	log.Printf("Processing %d keywords...", len(keywords))

	// 5. Process keywords sequentially
	var sessionIDs []string
	successCount := 0
	failedKeywords := []string{}

	for i, kw := range keywords {
		log.Printf("[%d/%d] Generating article for keyword: '%s'", i+1, len(keywords), kw.Keyword)

		// Generate article
		article, sessionID, err := ko.articleGen.GenerateArticle(ctx, kw.Keyword)
		sessionIDs = append(sessionIDs, sessionID)

		if err != nil {
			log.Printf("Error generating article for '%s': %v", kw.Keyword, err)
			failedKeywords = append(failedKeywords, kw.Keyword)
			// Continue processing other keywords
			continue
		}

		// Mark keyword as used
		if err := ko.keywordRepo.MarkAsUsed(kw.ID, article.ID); err != nil {
			log.Printf("Warning: Failed to mark keyword '%s' as used: %v", kw.Keyword, err)
			// Don't fail the whole job for this
		}

		successCount++
		log.Printf("✓ Successfully generated article: '%s' (ID: %s)", article.Title, article.ID)

		// Update job progress
		job.ArticlesGenerated = successCount
		job.ArticlesPublished = successCount // Auto-published
		job.SessionIDs = pq.StringArray(sessionIDs)
		ko.jobRepo.Update(job)
	}

	// 6. Mark job as completed
	completedAt := time.Now()
	job.Status = "completed"
	job.CompletedAt = &completedAt

	// Set error message if some failed
	if len(failedKeywords) > 0 {
		job.ErrorMessage = fmt.Sprintf("Failed keywords (%d/%d): %v", len(failedKeywords), len(keywords), failedKeywords)
	}

	ko.jobRepo.Update(job)

	log.Printf("Knowledge base update completed: %d/%d articles generated", successCount, len(keywords))
	if len(failedKeywords) > 0 {
		log.Printf("Warning: %d keywords failed: %v", len(failedKeywords), failedKeywords)
	}

	// 7. Check if keyword pool needs refilling after this batch
	go func() {
		refillCtx := context.Background()
		if err := ko.keywordPool.RefillPoolIfNeeded(refillCtx); err != nil {
			log.Printf("Warning: Post-update keyword pool refill failed: %v", err)
		}
	}()

	return job, nil
}

// GetJobStatus retrieves the current status of a job
func (ko *KBUpdateOrchestrator) GetJobStatus(ctx context.Context, jobID string) (*model.KBUpdateJob, error) {
	// Parse UUID
	id, err := uuid.Parse(jobID)
	if err != nil {
		return nil, fmt.Errorf("invalid job ID: %w", err)
	}
	return ko.jobRepo.GetByID(id)
}

// GetUpdateHistory retrieves job history with pagination
func (ko *KBUpdateOrchestrator) GetUpdateHistory(ctx context.Context, page, pageSize int) ([]model.KBUpdateJob, int64, error) {
	return ko.jobRepo.List(page, pageSize)
}
