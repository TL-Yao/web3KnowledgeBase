package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/llm"
	"github.com/user/web3-insight/internal/repository"
	"github.com/user/web3-insight/internal/service"
	"gorm.io/gorm"
)

type Server struct {
	config             *config.Config
	db                 *gorm.DB
	keyProvider        *service.KeyProvider
	articleHandler     *ArticleHandler
	categoryHandler    *CategoryHandler
	configHandler      *ConfigHandler
	taskHandler        *TaskHandler
	searchHandler      *SearchHandler
	chatHandler        *ChatHandler
	modelConfigHandler *ModelConfigHandler
	apiKeyHandler      *ApiKeyHandler
	tagHandler         *TagHandler
}

func NewServer(cfg *config.Config, db *gorm.DB, keyProvider *service.KeyProvider) *Server {
	// Initialize repositories
	articleRepo := repository.NewArticleRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	configRepo := repository.NewConfigRepository(db)
	taskRepo := repository.NewTaskRepository(db)

	// Build key functions from the shared KeyProvider
	claudeKeyFunc := func() string { return keyProvider.GetKey("anthropic") }
	openaiKeyFunc := func() string { return keyProvider.GetKey("openai") }

	// Initialize services
	chatService := service.NewChatService(db, &cfg.LLM, claudeKeyFunc, openaiKeyFunc)
	semanticSearchService := service.NewSemanticSearchService(articleRepo, &cfg.LLM)
	llmRouter := llm.NewRouterFromConfig(&cfg.LLM, claudeKeyFunc, openaiKeyFunc)
	articleUpdater := service.NewArticleUpdater(llmRouter)
	cliUpdater := service.NewCLIArticleUpdater()

	// Initialize tagger (shared across requests)
	tagRepo := repository.NewTagRepository(db)
	tagger := service.NewTagger(llmRouter, tagRepo, articleRepo, configRepo)

	return &Server{
		config:             cfg,
		db:                 db,
		keyProvider:        keyProvider,
		articleHandler:     NewArticleHandler(articleRepo, db, articleUpdater, cliUpdater),
		categoryHandler:    NewCategoryHandler(categoryRepo),
		configHandler:      NewConfigHandler(configRepo),
		taskHandler:        NewTaskHandler(taskRepo),
		searchHandler:      NewSearchHandlerWithSemantic(articleRepo, categoryRepo, semanticSearchService),
		chatHandler:        NewChatHandler(chatService),
		modelConfigHandler: NewModelConfigHandler(configRepo, cfg.Models, cfg.Routing),
		apiKeyHandler:      NewApiKeyHandler(configRepo, keyProvider),
		tagHandler:         NewTagHandler(db, tagger),
	}
}

func NewRouterWithDB(cfg *config.Config, db *gorm.DB, keyProvider *service.KeyProvider) *gin.Engine {
	router := gin.Default()

	// CORS middleware
	router.Use(corsMiddleware())

	// Initialize server with handlers
	server := NewServer(cfg, db, keyProvider)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// API routes
	api := router.Group("/api")
	{
		// Articles
		articles := api.Group("/articles")
		{
			articles.GET("", server.articleHandler.List)
			articles.GET("/:id", server.articleHandler.Get)
			articles.POST("", server.articleHandler.Create)
			articles.PUT("/:id", server.articleHandler.Update)
			articles.DELETE("/:id", server.articleHandler.Delete)
			articles.PATCH("/:id/archive", server.articleHandler.ToggleArchive)
			articles.PUT("/:id/tags", server.articleHandler.UpdateTags)
			articles.POST("/:id/generate-update", server.articleHandler.GenerateUpdate)
			articles.GET("/:id/update-status", server.articleHandler.GetUpdateStatus)
			articles.POST("/:id/cancel-update", server.articleHandler.CancelUpdate)
			articles.POST("/:id/apply-update", server.articleHandler.ApplyUpdate)
			articles.GET("/:id/versions", server.articleHandler.ListVersions)
			articles.POST("/:id/versions/:versionId/rollback", server.articleHandler.Rollback)
		}

		// Categories
		categories := api.Group("/categories")
		{
			categories.GET("", server.categoryHandler.List)
			categories.GET("/tree", server.categoryHandler.GetTree)
			categories.GET("/:id", server.categoryHandler.Get)
			categories.POST("", server.categoryHandler.Create)
			categories.PUT("/:id", server.categoryHandler.Update)
			categories.DELETE("/:id", server.categoryHandler.Delete)
		}

		// Search
		api.GET("/search", server.searchHandler.Search)
		api.GET("/search/semantic", server.searchHandler.SemanticSearch)

		// Related articles (under articles group would be better, but registered here for simplicity)
		articles.GET("/:id/related", server.searchHandler.RelatedArticles)

		// Config
		configGroup := api.Group("/config")
		{
			configGroup.GET("", server.configHandler.Get)
			configGroup.PUT("", server.configHandler.Update)
			configGroup.GET("/:key", server.configHandler.GetByKey)
			configGroup.PUT("/:key", server.configHandler.Set)
			configGroup.DELETE("/:key", server.configHandler.Delete)
		}

		// Tasks
		tasks := api.Group("/tasks")
		{
			tasks.GET("", server.taskHandler.List)
			tasks.GET("/stats", server.taskHandler.GetStats)
			tasks.GET("/:id", server.taskHandler.Get)
			tasks.POST("/:id/cancel", server.taskHandler.Cancel)
		}

		// Instant research (placeholder for now)
		api.POST("/research", func(c *gin.Context) {
			c.JSON(http.StatusAccepted, gin.H{
				"message": "research endpoint - to be implemented with LLM integration",
			})
		})

		// Data Sources
		dsHandler := NewDataSourceHandler(db)
		sources := api.Group("/sources")
		{
			sources.GET("", dsHandler.List)
			sources.GET("/:id", dsHandler.Get)
			sources.POST("", dsHandler.Create)
			sources.PUT("/:id", dsHandler.Update)
			sources.DELETE("/:id", dsHandler.Delete)
			sources.POST("/:id/sync", dsHandler.TriggerSync)
		}
		api.POST("/sources/validate", dsHandler.ValidateURL)

		// News Items
		newsHandler := NewNewsHandler(db)
		news := api.Group("/news")
		{
			news.GET("", newsHandler.List)
			news.GET("/unprocessed", newsHandler.GetUnprocessed)
			news.GET("/:id", newsHandler.Get)
			news.DELETE("/:id", newsHandler.Delete)
			news.POST("/:id/processed", newsHandler.MarkProcessed)
		}

		// Import/Export
		importHandler := NewImportHandler(db)
		importGroup := api.Group("/import")
		{
			importGroup.POST("", importHandler.Import)
			importGroup.POST("/validate", importHandler.Validate)
			importGroup.GET("/template", importHandler.GetTemplate)
			importGroup.GET("/export", importHandler.Export)
			importGroup.POST("/upload", importHandler.UploadFile)
		}

		// Explorer Research
		explorerHandler := NewExplorerHandler(db)
		explorers := api.Group("/explorers")
		{
			explorers.GET("", explorerHandler.List)
			explorers.GET("/chains", explorerHandler.GetChains)
			explorers.GET("/stats", explorerHandler.GetStats)
			explorers.GET("/features", explorerHandler.GetFeatures)
			explorers.POST("/features/seed", explorerHandler.SeedFeatures)
			explorers.GET("/compare", explorerHandler.Compare)
			explorers.GET("/:id", explorerHandler.Get)
			explorers.POST("", explorerHandler.Create)
			explorers.PUT("/:id", explorerHandler.Update)
			explorers.DELETE("/:id", explorerHandler.Delete)
			explorers.POST("/:id/status", explorerHandler.UpdateStatus)
		}

		// Model Configuration
		models := api.Group("/models")
		{
			models.GET("/registry", server.modelConfigHandler.GetModelsRegistry)
			models.GET("/tasks", server.modelConfigHandler.GetTaskTypes)
			models.GET("/selections", server.modelConfigHandler.GetUserSelections)
			models.PUT("/selections", server.modelConfigHandler.UpdateUserSelections)

			// API Key Management
			models.GET("/keys", server.apiKeyHandler.ListKeys)
			models.PUT("/keys", server.apiKeyHandler.SaveKeys)
			models.POST("/keys/test", server.apiKeyHandler.TestKey)
		}

		// Tags
		tags := api.Group("/tags")
		{
			tags.GET("", server.tagHandler.List)
			tags.POST("", server.tagHandler.Create)
			tags.GET("/search", server.tagHandler.Search)
			tags.GET("/in-use", server.tagHandler.GetInUse)
			tags.GET("/stats", server.tagHandler.GetStats)
			tags.PUT("/:id", server.tagHandler.Update)
			tags.DELETE("/:id", server.tagHandler.Delete)
			tags.PUT("/:id/status", server.tagHandler.UpdateStatus)
			tags.POST("/:id/approve", server.tagHandler.ApprovePending)
			tags.POST("/bulk-tag", server.tagHandler.BulkTag)
		}

		// Knowledge Base Update
		kbUpdateHandler := NewKBUpdateHandler(db, cfg, server.keyProvider)
		kb := api.Group("/kb")
		{
			// Update operations
			kb.POST("/update/trigger", kbUpdateHandler.TriggerUpdate)
			kb.GET("/update/jobs", kbUpdateHandler.GetUpdateHistory)
			kb.GET("/update/jobs/:job_id", kbUpdateHandler.GetJobStatus)

			// Keyword pool management
			kb.POST("/keywords/init", kbUpdateHandler.InitKeywordPool)
			kb.GET("/keywords/stats", kbUpdateHandler.GetKeywordStats)

			// Theme management
			kb.GET("/themes", kbUpdateHandler.GetThemes)
			kb.GET("/themes/active", kbUpdateHandler.GetActiveTheme)
			kb.PUT("/themes/:id/activate", kbUpdateHandler.SetActiveTheme)

			// KB config
			kb.GET("/config", kbUpdateHandler.GetKBConfig)
			kb.PUT("/config/batch-size", kbUpdateHandler.UpdateBatchSize)

			// Scheduler control
			kb.GET("/scheduler/status", kbUpdateHandler.GetSchedulerStatus)
			kb.POST("/scheduler/start", kbUpdateHandler.StartScheduler)
			kb.POST("/scheduler/stop", kbUpdateHandler.StopScheduler)
		}
	}

	// WebSocket for chat
	router.GET("/ws/chat", server.chatHandler.HandleWebSocket)

	return router
}
