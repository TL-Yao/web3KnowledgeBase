package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
)

const (
	DefaultKeywordBatchSize = 3  // Default number of keywords to process per update
	MaxKeywordBatchSize     = 10 // Maximum allowed batch size
)

type KBUpdateOrchestrator struct {
	keywordPool *KeywordPoolService
	articleGen  *ArticleGeneratorService
	keywordRepo *repository.KeywordRepository
	jobRepo     *repository.KBUpdateJobRepository
	themeRepo   *repository.ThemeRepository
	configRepo  *repository.ConfigRepository
	prompts     *config.PromptsConfig
}

func NewKBUpdateOrchestrator(
	keywordPool *KeywordPoolService,
	articleGen *ArticleGeneratorService,
	keywordRepo *repository.KeywordRepository,
	jobRepo *repository.KBUpdateJobRepository,
	themeRepo *repository.ThemeRepository,
	configRepo *repository.ConfigRepository,
	prompts *config.PromptsConfig,
) *KBUpdateOrchestrator {
	return &KBUpdateOrchestrator{
		keywordPool: keywordPool,
		articleGen:  articleGen,
		keywordRepo: keywordRepo,
		jobRepo:     jobRepo,
		themeRepo:   themeRepo,
		configRepo:  configRepo,
		prompts:     prompts,
	}
}

// GetBatchSize reads batch size from DB config, falling back to prompts.yaml default
func (ko *KBUpdateOrchestrator) GetBatchSize() int {
	if ko.configRepo != nil {
		cfg, err := ko.configRepo.Get("kb.batch_size")
		if err == nil {
			var val string
			if err := json.Unmarshal(cfg.Value, &val); err == nil {
				if size, err := strconv.Atoi(val); err == nil && size > 0 && size <= MaxKeywordBatchSize {
					return size
				}
			}
		}
	}
	if ko.prompts != nil && ko.prompts.Generation.DefaultBatchSize > 0 {
		return ko.prompts.Generation.DefaultBatchSize
	}
	return DefaultKeywordBatchSize
}

// RunUpdate executes a complete knowledge base update cycle
func (ko *KBUpdateOrchestrator) RunUpdate(ctx context.Context, triggerType string) (*model.KBUpdateJob, error) {
	log.Printf("Starting knowledge base update (trigger: %s)", triggerType)

	// 0. Get active theme
	activeTheme, err := ko.themeRepo.GetActive()
	if err != nil {
		return nil, fmt.Errorf("no active theme configured: %w", err)
	}
	themeID := activeTheme.ID
	log.Printf("Using active theme: %s (%s)", themeID, activeTheme.Name)

	// 1. Check for concurrent job execution
	runningJobs, err := ko.jobRepo.FindByStatus("running")
	if err != nil {
		return nil, fmt.Errorf("failed to check running jobs: %w", err)
	}

	if len(runningJobs) > 0 {
		return nil, fmt.Errorf("update already in progress (job: %s), please wait for completion", runningJobs[0].ID)
	}

	// 2. Cleanup orphaned jobs
	cleanedCount, err := ko.jobRepo.CleanupOrphanedJobs(30 * time.Minute)
	if err != nil {
		log.Printf("Warning: Failed to cleanup orphaned jobs: %v", err)
	} else if cleanedCount > 0 {
		log.Printf("Cleaned up %d orphaned jobs", cleanedCount)
	}

	// 3. Create job record
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

	// 4. Check and refill keyword pool for active theme
	log.Printf("Checking keyword pool for theme %s...", themeID)
	if err := ko.keywordPool.RefillPoolIfNeeded(ctx, themeID); err != nil {
		ko.jobRepo.UpdateStatus(job.ID, "failed", fmt.Sprintf("keyword pool refill failed: %v", err))
		return nil, fmt.Errorf("failed to refill keyword pool: %w", err)
	}

	// 5. Get pending keywords for active theme
	batchSize := ko.GetBatchSize()
	log.Printf("Fetching %d pending keywords for theme %s...", batchSize, themeID)
	keywords, err := ko.keywordRepo.GetPendingByTheme(themeID, batchSize)
	if err != nil {
		ko.jobRepo.UpdateStatus(job.ID, "failed", fmt.Sprintf("failed to fetch keywords: %v", err))
		return nil, fmt.Errorf("failed to get pending keywords: %w", err)
	}

	if len(keywords) < batchSize {
		errMsg := fmt.Sprintf("insufficient keywords for theme %s: got %d, need %d", themeID, len(keywords), batchSize)
		ko.jobRepo.UpdateStatus(job.ID, "failed", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	job.KeywordsGenerated = len(keywords)
	ko.jobRepo.Update(job)

	log.Printf("Processing %d keywords for theme %s...", len(keywords), themeID)

	// 6. Process keywords sequentially
	var sessionIDs []string
	successCount := 0
	failedKeywords := []string{}

	for i, kw := range keywords {
		log.Printf("[%d/%d] Generating article for keyword: '%s' (theme: %s)", i+1, len(keywords), kw.Keyword, themeID)

		article, sessionID, err := ko.articleGen.GenerateArticle(ctx, kw.Keyword, themeID)
		sessionIDs = append(sessionIDs, sessionID)

		if err != nil {
			log.Printf("Error generating article for '%s': %v", kw.Keyword, err)
			failedKeywords = append(failedKeywords, kw.Keyword)
			continue
		}

		if err := ko.keywordRepo.MarkAsUsed(kw.ID, article.ID); err != nil {
			log.Printf("Warning: Failed to mark keyword '%s' as used: %v", kw.Keyword, err)
		}

		successCount++
		log.Printf("Successfully generated article: '%s' (ID: %s)", article.Title, article.ID)

		job.ArticlesGenerated = successCount
		job.ArticlesPublished = successCount
		job.SessionIDs = pq.StringArray(sessionIDs)
		ko.jobRepo.Update(job)
	}

	// 7. Mark job as completed
	completedAt := time.Now()
	job.Status = "completed"
	job.CompletedAt = &completedAt

	if len(failedKeywords) > 0 {
		job.ErrorMessage = fmt.Sprintf("Failed keywords (%d/%d): %v", len(failedKeywords), len(keywords), failedKeywords)
	}

	ko.jobRepo.Update(job)

	log.Printf("Knowledge base update completed: %d/%d articles generated", successCount, len(keywords))
	if len(failedKeywords) > 0 {
		log.Printf("Warning: %d keywords failed: %v", len(failedKeywords), failedKeywords)
	}

	// 8. Post-update: check if keyword pool needs refilling
	go func() {
		refillCtx := context.Background()
		if err := ko.keywordPool.RefillPoolIfNeeded(refillCtx, themeID); err != nil {
			log.Printf("Warning: Post-update keyword pool refill failed: %v", err)
		}
	}()

	return job, nil
}

// GetJobStatus retrieves the current status of a job
func (ko *KBUpdateOrchestrator) GetJobStatus(ctx context.Context, jobID string) (*model.KBUpdateJob, error) {
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
