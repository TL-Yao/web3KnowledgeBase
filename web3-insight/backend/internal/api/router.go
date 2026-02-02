package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/repository"
	"github.com/user/web3-insight/internal/service"
	"gorm.io/gorm"
)

type Server struct {
	config          *config.Config
	db              *gorm.DB
	articleHandler  *ArticleHandler
	categoryHandler *CategoryHandler
	configHandler   *ConfigHandler
	taskHandler     *TaskHandler
	searchHandler   *SearchHandler
	chatHandler     *ChatHandler
}

func NewServer(cfg *config.Config, db *gorm.DB) *Server {
	// Initialize repositories
	articleRepo := repository.NewArticleRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	configRepo := repository.NewConfigRepository(db)
	taskRepo := repository.NewTaskRepository(db)

	// Initialize services
	chatService := service.NewChatService(db, &cfg.LLM)
	semanticSearchService := service.NewSemanticSearchService(articleRepo, &cfg.LLM)

	return &Server{
		config:          cfg,
		db:              db,
		articleHandler:  NewArticleHandler(articleRepo),
		categoryHandler: NewCategoryHandler(categoryRepo),
		configHandler:   NewConfigHandler(configRepo),
		taskHandler:     NewTaskHandler(taskRepo),
		searchHandler:   NewSearchHandlerWithSemantic(articleRepo, categoryRepo, semanticSearchService),
		chatHandler:     NewChatHandler(chatService),
	}
}

func NewRouter(cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// CORS middleware
	router.Use(corsMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	return router
}

func NewRouterWithDB(cfg *config.Config, db *gorm.DB) *gin.Engine {
	router := gin.Default()

	// CORS middleware
	router.Use(corsMiddleware())

	// Initialize server with handlers
	server := NewServer(cfg, db)

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
			articles.POST("/:id/regenerate", server.articleHandler.Regenerate)
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
	}

	// WebSocket for chat
	router.GET("/ws/chat", server.chatHandler.HandleWebSocket)

	return router
}
