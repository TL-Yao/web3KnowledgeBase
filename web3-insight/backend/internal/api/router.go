package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/user/web3-insight/internal/config"
)

func NewRouter(cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Articles
		v1.GET("/articles", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "list articles"})
		})

		// Categories
		v1.GET("/categories", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "list categories"})
		})

		// Chat
		v1.POST("/chat", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "chat endpoint"})
		})
	}

	return router
}
