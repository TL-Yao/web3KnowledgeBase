package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/user/web3-insight/internal/repository"
)

type ConfigHandler struct {
	repo *repository.ConfigRepository
}

func NewConfigHandler(repo *repository.ConfigRepository) *ConfigHandler {
	return &ConfigHandler{repo: repo}
}

// GetConfig godoc
// @Summary Get all configuration values
// @Description Get all configuration key-value pairs
// @Tags config
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/config [get]
func (h *ConfigHandler) Get(c *gin.Context) {
	configs, err := h.repo.GetMap()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, configs)
}

// GetConfigByKey godoc
// @Summary Get configuration value by key
// @Description Get a single configuration value by its key
// @Tags config
// @Accept json
// @Produce json
// @Param key path string true "Config key"
// @Success 200 {object} model.Config
// @Router /api/config/{key} [get]
func (h *ConfigHandler) GetByKey(c *gin.Context) {
	key := c.Param("key")

	config, err := h.repo.Get(key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "config not found"})
		return
	}

	c.JSON(http.StatusOK, config)
}

type UpdateConfigRequest struct {
	Configs map[string]string `json:"configs" binding:"required"`
}

// UpdateConfig godoc
// @Summary Update configuration values
// @Description Update multiple configuration values at once
// @Tags config
// @Accept json
// @Produce json
// @Param config body UpdateConfigRequest true "Configuration values"
// @Success 200 {object} map[string]string
// @Router /api/config [put]
func (h *ConfigHandler) Update(c *gin.Context) {
	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.SetMultiple(req.Configs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	configs, err := h.repo.GetMap()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, configs)
}

type SetConfigRequest struct {
	Value       string `json:"value" binding:"required"`
	Description string `json:"description"`
}

// SetConfig godoc
// @Summary Set a single configuration value
// @Description Set a configuration value by key
// @Tags config
// @Accept json
// @Produce json
// @Param key path string true "Config key"
// @Param config body SetConfigRequest true "Configuration value"
// @Success 200 {object} model.Config
// @Router /api/config/{key} [put]
func (h *ConfigHandler) Set(c *gin.Context) {
	key := c.Param("key")

	var req SetConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Set(key, req.Value, req.Description); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	config, err := h.repo.Get(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// DeleteConfig godoc
// @Summary Delete a configuration value
// @Description Delete a configuration value by key
// @Tags config
// @Accept json
// @Produce json
// @Param key path string true "Config key"
// @Success 204 "No Content"
// @Router /api/config/{key} [delete]
func (h *ConfigHandler) Delete(c *gin.Context) {
	key := c.Param("key")

	if err := h.repo.Delete(key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
