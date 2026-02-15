package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/gosimple/slug"
	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
)

// Research-specific delimiters for structured output parsing
const (
	PlanStart          = "===PLAN_START==="
	PlanEnd            = "===PLAN_END==="
	ReportTitleStart   = "===REPORT_TITLE_START==="
	ReportTitleEnd     = "===REPORT_TITLE_END==="
	ReportContentStart = "===REPORT_CONTENT_START==="
	ReportContentEnd   = "===REPORT_CONTENT_END==="
	ReportSummaryStart = "===REPORT_SUMMARY_START==="
	ReportSummaryEnd   = "===REPORT_SUMMARY_END==="
	CitationsStart     = "===CITATIONS_START==="
	CitationsEnd       = "===CITATIONS_END==="
)

const maxConcurrentSessions = 3

// StartResearchRequest contains the input for starting a research session.
type StartResearchRequest struct {
	Question   string `json:"question"`
	Domain     string `json:"domain"`
	ReviewPlan bool   `json:"reviewPlan"`
}

// PinnedFinding represents a single pinned chat finding.
type PinnedFinding struct {
	MessageContent string    `json:"messageContent"`
	PinnedAt       time.Time `json:"pinnedAt"`
	Integrated     bool      `json:"integrated"`
}

// CitationEntry represents a single citation extracted from the report.
type CitationEntry struct {
	URL        string `json:"url"`
	Title      string `json:"title"`
	Snippet    string `json:"snippet"`
	AccessedAt string `json:"accessedAt"`
}

type ResearchService struct {
	sessionRepo    *repository.ResearchSessionRepository
	articleRepo    *repository.ArticleRepository
	researchConfig *config.ResearchConfig
	cancelFuncs    sync.Map // map[uuid.UUID]context.CancelFunc
}

func NewResearchService(
	sessionRepo *repository.ResearchSessionRepository,
	articleRepo *repository.ArticleRepository,
	researchConfig *config.ResearchConfig,
) *ResearchService {
	return &ResearchService{
		sessionRepo:    sessionRepo,
		articleRepo:    articleRepo,
		researchConfig: researchConfig,
	}
}

// StartSession creates a research session and kicks off plan generation in a goroutine.
func (s *ResearchService) StartSession(ctx context.Context, req StartResearchRequest) (*model.ResearchSession, error) {
	// Check concurrent session limit
	activeCount, err := s.sessionRepo.CountActive()
	if err != nil {
		return nil, fmt.Errorf("failed to check active sessions: %w", err)
	}
	if activeCount >= maxConcurrentSessions {
		return nil, fmt.Errorf("concurrent session limit reached (%d)", maxConcurrentSessions)
	}

	// Validate domain
	if req.Domain == "" {
		req.Domain = "auto"
	}
	if s.researchConfig != nil {
		if _, err := s.researchConfig.GetDomainByID(req.Domain); err != nil {
			return nil, fmt.Errorf("invalid domain: %s", req.Domain)
		}
	}

	now := time.Now()
	session := &model.ResearchSession{
		Question:    req.Question,
		Domain:      req.Domain,
		Status:      "planning",
		Stage:       "planning",
		StageDetail: "正在分析问题...",
		StartedAt:   &now,
	}

	if err := s.sessionRepo.Create(session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Run generation pipeline in background goroutine with cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFuncs.Store(session.ID, cancel)
	go func() {
		defer s.cancelFuncs.Delete(session.ID)
		s.runPlanGeneration(ctx, session.ID, req)
	}()

	return session, nil
}

// ApprovePlan moves session from plan_review → researching and starts research generation.
func (s *ResearchService) ApprovePlan(ctx context.Context, sessionID uuid.UUID, editedPlan string) error {
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	if session.Status != "plan_review" {
		return fmt.Errorf("session is not in plan_review status (current: %s)", session.Status)
	}

	// Update plan if edited
	if editedPlan != "" {
		session.ResearchPlan = editedPlan
	}
	session.PlanApproved = true
	if err := s.sessionRepo.Update(session); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	// Continue to research phase in background with cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFuncs.Store(sessionID, cancel)
	go func() {
		defer s.cancelFuncs.Delete(sessionID)
		s.runResearchGeneration(ctx, sessionID)
	}()

	return nil
}

// CancelSession cancels an in-progress research session.
func (s *ResearchService) CancelSession(ctx context.Context, sessionID uuid.UUID) error {
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	activeStatuses := map[string]bool{
		"pending": true, "planning": true, "plan_review": true,
		"researching": true, "writing": true,
	}
	if !activeStatuses[session.Status] {
		return fmt.Errorf("session cannot be cancelled (status: %s)", session.Status)
	}

	// Cancel the running goroutine if active
	if cancelFn, ok := s.cancelFuncs.LoadAndDelete(sessionID); ok {
		cancelFn.(context.CancelFunc)()
		log.Printf("Cancelled running goroutine for session %s", sessionID)
	}

	return s.sessionRepo.UpdateStatus(sessionID, "failed", "", "Cancelled by user")
}

// PinFinding adds a chat message content to the session's pinned findings.
func (s *ResearchService) PinFinding(ctx context.Context, sessionID uuid.UUID, content string) error {
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	var findings []PinnedFinding
	if len(session.PinnedFindings) > 0 {
		if err := json.Unmarshal(session.PinnedFindings, &findings); err != nil {
			findings = []PinnedFinding{}
		}
	}

	findings = append(findings, PinnedFinding{
		MessageContent: content,
		PinnedAt:       time.Now(),
		Integrated:     false,
	})

	data, err := json.Marshal(findings)
	if err != nil {
		return fmt.Errorf("failed to marshal pinned findings: %w", err)
	}

	return s.sessionRepo.UpdatePinnedFindings(sessionID, data)
}

// RemovePinFinding removes a pinned finding by index.
func (s *ResearchService) RemovePinFinding(ctx context.Context, sessionID uuid.UUID, index int) error {
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	var findings []PinnedFinding
	if err := json.Unmarshal(session.PinnedFindings, &findings); err != nil {
		return fmt.Errorf("failed to parse pinned findings: %w", err)
	}

	if index < 0 || index >= len(findings) {
		return fmt.Errorf("invalid finding index: %d", index)
	}

	findings = append(findings[:index], findings[index+1:]...)

	data, err := json.Marshal(findings)
	if err != nil {
		return fmt.Errorf("failed to marshal pinned findings: %w", err)
	}

	return s.sessionRepo.UpdatePinnedFindings(sessionID, data)
}

// IntegrateFindings re-runs the writing stage to merge pinned chat content into the report.
func (s *ResearchService) IntegrateFindings(ctx context.Context, sessionID uuid.UUID) error {
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	if session.Status != "completed" {
		return fmt.Errorf("can only integrate findings for completed sessions (current: %s)", session.Status)
	}

	// Check concurrent limit
	activeCount, err := s.sessionRepo.CountActive()
	if err != nil {
		return fmt.Errorf("failed to check active sessions: %w", err)
	}
	if activeCount >= maxConcurrentSessions {
		return fmt.Errorf("concurrent session limit reached (%d)", maxConcurrentSessions)
	}

	// Set status back to writing
	if err := s.sessionRepo.UpdateStatus(sessionID, "writing", "writing", "正在整合研究发现..."); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFuncs.Store(sessionID, cancel)
	go func() {
		defer s.cancelFuncs.Delete(sessionID)
		s.runIntegration(ctx, sessionID)
	}()

	return nil
}

// CleanupOrphanedSessions marks stuck sessions as failed. Called on startup.
func (s *ResearchService) CleanupOrphanedSessions() {
	cleaned, err := s.sessionRepo.CleanupOrphaned(45 * time.Minute)
	if err != nil {
		log.Printf("Warning: failed to cleanup orphaned research sessions: %v", err)
		return
	}
	if cleaned > 0 {
		log.Printf("Cleaned up %d orphaned research sessions", cleaned)
	}
}

// --- Internal pipeline methods ---

func (s *ResearchService) runPlanGeneration(ctx context.Context, sessionID uuid.UUID, req StartResearchRequest) {
	planTimeout := 3 * time.Minute
	if s.researchConfig != nil && s.researchConfig.Generation.PlanTimeoutMinutes > 0 {
		planTimeout = time.Duration(s.researchConfig.Generation.PlanTimeoutMinutes) * time.Minute
	}

	planCtx, cancel := context.WithTimeout(ctx, planTimeout)
	defer cancel()

	// Build domain context
	domainContext := ""
	if s.researchConfig != nil {
		if domain, err := s.researchConfig.GetDomainByID(req.Domain); err == nil {
			domainContext = domain.SystemContext
		}
	}

	// Build plan prompt
	prompt, err := s.buildPlanPrompt(req.Question, domainContext)
	if err != nil {
		s.failSession(sessionID, fmt.Sprintf("Failed to build plan prompt: %v", err))
		return
	}

	// Create executor with Opus model and subscription auth
	executor := NewClaudeExecutorWithOptions(ClaudeExecutorOptions{
		StripAPIKey: true,
		Model:       s.getModel(),
	})

	log.Printf("Research plan generation started (session: %s, cli: %s)", sessionID, executor.GetSessionID())

	response, err := executor.Execute(planCtx, prompt)
	if err != nil {
		s.failSession(sessionID, fmt.Sprintf("Plan generation failed: %v", err))
		return
	}

	// Parse plan
	plan, err := extractBetween(response.Result, PlanStart, PlanEnd)
	if err != nil {
		// If delimiters not found, use the entire response as the plan
		log.Printf("Plan delimiter extraction failed, using full response: %v", err)
		plan = response.Result
	}
	plan = strings.TrimSpace(plan)

	// Update session with plan and cost
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		s.failSession(sessionID, fmt.Sprintf("Failed to load session: %v", err))
		return
	}
	session.ResearchPlan = plan
	session.GenerationCost = response.TotalCostUSD
	session.CLISessionID = executor.GetSessionID()

	if req.ReviewPlan {
		// Pause for user review
		session.Status = "plan_review"
		session.Stage = "plan_review"
		session.StageDetail = "Research plan ready for review"
		if err := s.sessionRepo.Update(session); err != nil {
			log.Printf("Failed to update session for plan review: %v", err)
		}
		log.Printf("Research plan ready for review (session: %s)", sessionID)
		return
	}

	// Auto-approve and continue
	session.PlanApproved = true
	if err := s.sessionRepo.Update(session); err != nil {
		s.failSession(sessionID, fmt.Sprintf("Failed to update session: %v", err))
		return
	}

	s.runResearchGeneration(ctx, sessionID)
}

func (s *ResearchService) runResearchGeneration(ctx context.Context, sessionID uuid.UUID) {
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		s.failSession(sessionID, fmt.Sprintf("Failed to load session: %v", err))
		return
	}

	// Update status to researching (stays during the entire CLI execution)
	s.sessionRepo.UpdateStatus(sessionID, "researching", "researching", "正在搜索资料...")

	researchTimeout := 30 * time.Minute
	if s.researchConfig != nil && s.researchConfig.Generation.TimeoutMinutes > 0 {
		researchTimeout = time.Duration(s.researchConfig.Generation.TimeoutMinutes) * time.Minute
	}

	researchCtx, cancel := context.WithTimeout(ctx, researchTimeout)
	defer cancel()

	// Build domain context
	domainContext := ""
	if s.researchConfig != nil {
		if domain, err := s.researchConfig.GetDomainByID(session.Domain); err == nil {
			domainContext = domain.SystemContext
		}
	}

	// Build research prompt
	prompt, err := s.buildResearchPrompt(session.Question, session.ResearchPlan, domainContext, nil)
	if err != nil {
		s.failSession(sessionID, fmt.Sprintf("Failed to build research prompt: %v", err))
		return
	}

	// Create new executor (fresh session ID)
	executor := NewClaudeExecutorWithOptions(ClaudeExecutorOptions{
		StripAPIKey: true,
		Model:       s.getModel(),
	})

	log.Printf("Research generation started (session: %s, cli: %s)", sessionID, executor.GetSessionID())

	response, err := executor.Execute(researchCtx, prompt)
	if err != nil {
		s.failSession(sessionID, fmt.Sprintf("Research generation failed: %v", err))
		return
	}

	// Update stage to writing/finalizing as we parse and save the result
	s.sessionRepo.UpdateStatus(sessionID, "writing", "writing", "正在撰写报告...")

	// Parse response and save
	s.saveResearchResult(sessionID, session, response)
}

func (s *ResearchService) runIntegration(ctx context.Context, sessionID uuid.UUID) {
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		s.failSession(sessionID, fmt.Sprintf("Failed to load session: %v", err))
		return
	}

	// Load existing article content
	var existingContent string
	if session.ArticleID != nil {
		article, err := s.articleRepo.GetByID(*session.ArticleID)
		if err == nil {
			existingContent = article.Content
		}
	}

	// Parse pinned findings
	var findings []PinnedFinding
	if len(session.PinnedFindings) > 0 {
		json.Unmarshal(session.PinnedFindings, &findings)
	}

	// Filter to unintegrated findings only
	var newFindings []PinnedFinding
	for _, f := range findings {
		if !f.Integrated {
			newFindings = append(newFindings, f)
		}
	}
	if len(newFindings) == 0 {
		s.sessionRepo.UpdateStatus(sessionID, "completed", "", "")
		return
	}

	// Build domain context
	domainContext := ""
	if s.researchConfig != nil {
		if domain, err := s.researchConfig.GetDomainByID(session.Domain); err == nil {
			domainContext = domain.SystemContext
		}
	}

	integrationTimeout := 30 * time.Minute
	if s.researchConfig != nil && s.researchConfig.Generation.TimeoutMinutes > 0 {
		integrationTimeout = time.Duration(s.researchConfig.Generation.TimeoutMinutes) * time.Minute
	}

	intCtx, cancel := context.WithTimeout(ctx, integrationTimeout)
	defer cancel()

	prompt, err := s.buildIntegrationPrompt(session.Question, existingContent, domainContext, newFindings)
	if err != nil {
		s.failSession(sessionID, fmt.Sprintf("Failed to build integration prompt: %v", err))
		return
	}

	executor := NewClaudeExecutorWithOptions(ClaudeExecutorOptions{
		StripAPIKey: true,
		Model:       s.getModel(),
	})

	log.Printf("Research integration started (session: %s, findings: %d)", sessionID, len(newFindings))

	response, err := executor.Execute(intCtx, prompt)
	if err != nil {
		s.failSession(sessionID, fmt.Sprintf("Integration failed: %v", err))
		return
	}

	// Mark findings as integrated
	for i := range findings {
		findings[i].Integrated = true
	}
	findingsData, _ := json.Marshal(findings)
	s.sessionRepo.UpdatePinnedFindings(sessionID, findingsData)

	// Save updated result
	s.saveResearchResult(sessionID, session, response)
}

func (s *ResearchService) saveResearchResult(sessionID uuid.UUID, session *model.ResearchSession, response *ClaudeResponse) {
	// Parse title
	title, err := extractBetween(response.Result, ReportTitleStart, ReportTitleEnd)
	if err != nil {
		// Fallback: use question as title
		title = session.Question
		if len(title) > 100 {
			title = title[:100]
		}
	}
	title = strings.TrimSpace(title)

	// Parse content
	content, err := extractBetween(response.Result, ReportContentStart, ReportContentEnd)
	if err != nil {
		s.failSession(sessionID, fmt.Sprintf("Failed to extract report content: %v", err))
		return
	}
	content = strings.TrimSpace(content)

	// Validate content length
	minLength := 1000
	if s.researchConfig != nil && s.researchConfig.Generation.MinReportLength > 0 {
		minLength = s.researchConfig.Generation.MinReportLength
	}
	if len(content) < minLength {
		s.failSession(sessionID, fmt.Sprintf("Report too short: %d chars (min: %d)", len(content), minLength))
		return
	}

	// Parse summary (optional)
	summary, _ := extractBetween(response.Result, ReportSummaryStart, ReportSummaryEnd)
	summary = strings.TrimSpace(summary)

	// Parse citations (optional)
	var citations []CitationEntry
	citationsRaw, err := extractBetween(response.Result, CitationsStart, CitationsEnd)
	if err == nil {
		json.Unmarshal([]byte(strings.TrimSpace(citationsRaw)), &citations)
	}

	// Extract source URLs from citations
	var sourceURLs []string
	for _, c := range citations {
		if c.URL != "" {
			sourceURLs = append(sourceURLs, c.URL)
		}
	}

	// Store citations JSON on session
	citationsJSON, _ := json.Marshal(citations)

	// Reload session to get latest state
	session, err = s.sessionRepo.GetByID(sessionID)
	if err != nil {
		s.failSession(sessionID, fmt.Sprintf("Failed to reload session: %v", err))
		return
	}

	// Update session cost
	session.GenerationCost += response.TotalCostUSD
	session.Citations = citationsJSON

	// Create or update article
	if session.ArticleID != nil {
		// Update existing article (integration case)
		article, err := s.articleRepo.GetByID(*session.ArticleID)
		if err != nil {
			s.failSession(sessionID, fmt.Sprintf("Failed to load article: %v", err))
			return
		}
		article.Title = title
		article.Content = content
		article.ContentHTML = "" // Clear stale HTML
		article.Summary = summary
		article.SourceURLs = sourceURLs
		if err := s.articleRepo.Update(article); err != nil {
			s.failSession(sessionID, fmt.Sprintf("Failed to update article: %v", err))
			return
		}
		log.Printf("Research article updated: ID=%s", article.ID)
	} else {
		// Create new article
		articleSlug := s.generateSlug(title)
		tags := []string{"research"}
		if session.Domain != "auto" {
			tags = append(tags, session.Domain)
		}

		article := &model.Article{
			Title:       title,
			Slug:        articleSlug,
			Content:     content,
			Summary:     summary,
			Status:      "published",
			ArticleType: "research",
			Tags:        tags,
			SourceURLs:  sourceURLs,
		}

		if err := s.articleRepo.Create(article); err != nil {
			s.failSession(sessionID, fmt.Sprintf("Failed to save article: %v", err))
			return
		}

		// Link article to session
		if err := s.sessionRepo.SetArticleID(sessionID, article.ID); err != nil {
			log.Printf("Warning: failed to link article to session: %v", err)
		}
		session.ArticleID = &article.ID
		log.Printf("Research article created: ID=%s, Slug=%s", article.ID, articleSlug)
	}

	// Mark session as completed
	now := time.Now()
	session.Status = "completed"
	session.Stage = ""
	session.StageDetail = ""
	session.CompletedAt = &now
	if err := s.sessionRepo.Update(session); err != nil {
		log.Printf("Warning: failed to mark session as completed: %v", err)
	}

	log.Printf("Research session completed (session: %s, cost: $%.4f)", sessionID, session.GenerationCost)
}

// --- Prompt builders ---

func (s *ResearchService) buildPlanPrompt(question, domainContext string) (string, error) {
	tmplStr := `You are a research planning assistant. Create a detailed research outline for the following question.

{{if .DomainContext}}## Domain Context
{{.DomainContext}}
{{end}}
## Research Question
{{.Question}}

## Instructions
1. Analyze the question and identify key aspects to research
2. Create a structured outline with 4-8 main sections
3. For each section, list 2-4 specific points to investigate
4. Identify potential sources and search queries
5. Note any specific data points or comparisons needed

## Output Format
Wrap your research plan in these exact delimiters:

===PLAN_START===
Your research plan here (markdown format)
===PLAN_END===`

	tmpl, err := template.New("plan").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]interface{}{
		"Question":      question,
		"DomainContext": domainContext,
	})
	return buf.String(), err
}

func (s *ResearchService) buildResearchPrompt(question, plan, domainContext string, findings []PinnedFinding) (string, error) {
	tmplStr := `You are an expert research analyst. Generate a comprehensive, factual research report based on the following plan.

{{if .DomainContext}}## Domain Context
{{.DomainContext}}
{{end}}
## Research Question
{{.Question}}

## Research Plan
{{.Plan}}

{{if .Findings}}## Additional Findings to Incorporate
The user has pinned these additional findings from chat. Integrate them naturally into the report:
{{range .Findings}}
---
{{.MessageContent}}
---
{{end}}
{{end}}
## Critical Instructions
1. Use web search and reading tools to gather FACTUAL, up-to-date information
2. EVERY claim must be backed by a source. Use inline citations [1], [2], etc.
3. Be comprehensive: minimum 1500 words, cover all sections in the plan
4. Language: Write in Chinese (中文). Use "English (中文)" for technical terms
5. Include specific data, numbers, comparisons where available
6. Be objective: present multiple perspectives where relevant
7. DO NOT hallucinate or make up facts. If information is unavailable, say so.

## Output Format (use these EXACT delimiters)

===REPORT_TITLE_START===
Report title in Chinese
===REPORT_TITLE_END===
===REPORT_CONTENT_START===
Full report content in Markdown with inline citations [1], [2], etc.
===REPORT_CONTENT_END===
===REPORT_SUMMARY_START===
80-150 word summary in Chinese
===REPORT_SUMMARY_END===
===CITATIONS_START===
[{"url":"https://...","title":"Source title","snippet":"Brief description","accessedAt":"2026-02-15"}]
===CITATIONS_END===

IMPORTANT: Citations must be valid JSON array. Include ALL sources used.`

	tmpl, err := template.New("research").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]interface{}{
		"Question":      question,
		"Plan":          plan,
		"DomainContext": domainContext,
		"Findings":      findings,
	})
	return buf.String(), err
}

func (s *ResearchService) buildIntegrationPrompt(question, existingContent, domainContext string, findings []PinnedFinding) (string, error) {
	tmplStr := `You are an expert research analyst. Update and improve the following research report by integrating new findings from the user's chat research.

{{if .DomainContext}}## Domain Context
{{.DomainContext}}
{{end}}
## Research Question
{{.Question}}

## Existing Report Content
{{.ExistingContent}}

## New Findings to Integrate
Integrate these findings naturally into the report. Add new sections if needed, update existing sections with new information, and add new citations.
{{range .Findings}}
---
{{.MessageContent}}
---
{{end}}

## Instructions
1. Preserve the existing report structure and content
2. Seamlessly integrate the new findings in appropriate sections
3. Add new citations for any new sources
4. Maintain consistent tone and language
5. Keep or improve inline citation numbering [1], [2], etc.
6. Language: Chinese (中文)

## Output Format (use these EXACT delimiters)

===REPORT_TITLE_START===
Updated report title in Chinese (can keep original if still appropriate)
===REPORT_TITLE_END===
===REPORT_CONTENT_START===
Full updated report content in Markdown with citations
===REPORT_CONTENT_END===
===REPORT_SUMMARY_START===
Updated 80-150 word summary in Chinese
===REPORT_SUMMARY_END===
===CITATIONS_START===
[{"url":"https://...","title":"Source title","snippet":"Brief description","accessedAt":"2026-02-15"}]
===CITATIONS_END===`

	tmpl, err := template.New("integration").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]interface{}{
		"Question":        question,
		"ExistingContent": existingContent,
		"DomainContext":   domainContext,
		"Findings":        findings,
	})
	return buf.String(), err
}

// --- Helpers ---

func (s *ResearchService) failSession(sessionID uuid.UUID, errMsg string) {
	log.Printf("Research session failed (session: %s): %s", sessionID, errMsg)
	s.sessionRepo.UpdateStatus(sessionID, "failed", "", errMsg)
}

func (s *ResearchService) getModel() string {
	if s.researchConfig != nil && s.researchConfig.Generation.DefaultModel != "" {
		return s.researchConfig.Generation.DefaultModel
	}
	return "opus"
}

func (s *ResearchService) generateSlug(title string) string {
	baseSlug := slug.Make(title)
	if baseSlug == "" {
		baseSlug = fmt.Sprintf("research-%d", time.Now().Unix())
	}

	existingCount := s.articleRepo.CountBySlugPrefix(baseSlug)
	if existingCount > 0 {
		baseSlug = fmt.Sprintf("%s-%d", baseSlug, existingCount+1)
	}

	return baseSlug
}
