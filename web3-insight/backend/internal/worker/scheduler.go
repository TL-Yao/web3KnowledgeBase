package worker

import (
	"log"

	"github.com/hibiken/asynq"
)

// Scheduler manages periodic task scheduling
type Scheduler struct {
	scheduler *asynq.Scheduler
	client    *asynq.Client
}

// NewScheduler creates a new task scheduler
func NewScheduler(redisOpt asynq.RedisClientOpt) *Scheduler {
	return &Scheduler{
		scheduler: asynq.NewScheduler(redisOpt, nil),
		client:    asynq.NewClient(redisOpt),
	}
}

// RegisterTasks registers all periodic tasks
func (s *Scheduler) RegisterTasks() error {
	var err error

	// RSS sync every hour
	task, _ := NewRSSSyncTask(RSSSyncPayload{})
	_, err = s.scheduler.Register("0 * * * *", task, asynq.Queue("default"))
	if err != nil {
		log.Printf("Failed to register RSS sync task: %v", err)
		return err
	}
	log.Println("Registered RSS sync task: every hour")

	// Content generation every 6 hours (for suggested topics)
	task, _ = NewContentGenerateTask(ContentGeneratePayload{
		Topic: "suggested",
		Style: "auto",
	})
	_, err = s.scheduler.Register("0 */6 * * *", task, asynq.Queue("low"))
	if err != nil {
		log.Printf("Failed to register content generation task: %v", err)
		return err
	}
	log.Println("Registered content generation task: every 6 hours")

	return nil
}

// Run starts the scheduler
func (s *Scheduler) Run() error {
	log.Println("Scheduler starting...")
	return s.scheduler.Run()
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.scheduler.Shutdown()
	s.client.Close()
}

// EnqueueTask enqueues a task for immediate processing
func (s *Scheduler) EnqueueTask(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return s.client.Enqueue(task, opts...)
}

// EnqueueContentGenerate enqueues a content generation task
func (s *Scheduler) EnqueueContentGenerate(topic, categoryID, style string) (*asynq.TaskInfo, error) {
	task, err := NewContentGenerateTask(ContentGeneratePayload{
		Topic:      topic,
		CategoryID: categoryID,
		Style:      style,
	})
	if err != nil {
		return nil, err
	}
	return s.client.Enqueue(task, asynq.Queue("default"))
}

// EnqueueRSSSync enqueues an RSS sync task
func (s *Scheduler) EnqueueRSSSync(feedURL, categoryID string) (*asynq.TaskInfo, error) {
	task, err := NewRSSSyncTask(RSSSyncPayload{
		FeedURL:    feedURL,
		CategoryID: categoryID,
	})
	if err != nil {
		return nil, err
	}
	return s.client.Enqueue(task, asynq.Queue("default"))
}

// EnqueueWebCrawl enqueues a web crawl task
func (s *Scheduler) EnqueueWebCrawl(url, categoryID string, depth int) (*asynq.TaskInfo, error) {
	task, err := NewWebCrawlTask(WebCrawlPayload{
		URL:        url,
		CategoryID: categoryID,
		Depth:      depth,
	})
	if err != nil {
		return nil, err
	}
	return s.client.Enqueue(task, asynq.Queue("default"))
}

// EnqueueClassify enqueues a classification task
func (s *Scheduler) EnqueueClassify(articleID string) (*asynq.TaskInfo, error) {
	task, err := NewClassifyTask(ClassifyPayload{
		ArticleID: articleID,
	})
	if err != nil {
		return nil, err
	}
	return s.client.Enqueue(task, asynq.Queue("default"))
}

// EnqueueEmbedding enqueues an embedding generation task
func (s *Scheduler) EnqueueEmbedding(articleID string) (*asynq.TaskInfo, error) {
	task, err := NewEmbeddingTask(EmbeddingPayload{
		ArticleID: articleID,
	})
	if err != nil {
		return nil, err
	}
	return s.client.Enqueue(task, asynq.Queue("default"))
}
