package api

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/repository"
	"github.com/user/web3-insight/internal/service"
	"gorm.io/gorm"
)

type ImportHandler struct {
	importer *service.ArticleImporter
}

func NewImportHandler(db *gorm.DB) *ImportHandler {
	articleRepo := repository.NewArticleRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	return &ImportHandler{
		importer: service.NewArticleImporter(articleRepo, categoryRepo),
	}
}

// Import godoc
// @Summary Import articles from JSON
// @Description Import multiple articles from JSON format
// @Tags import
// @Accept json
// @Produce json
// @Param body body service.ImportBatch true "Import batch"
// @Success 200 {object} service.ImportResult
// @Failure 400 {object} map[string]string
// @Router /api/import [post]
func (h *ImportHandler) Import(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	batch, err := h.importer.ParseJSON(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(batch.Articles) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no articles to import"})
		return
	}

	result, err := h.importer.Import(*batch)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Validate godoc
// @Summary Validate import data
// @Description Validate import data without actually importing
// @Tags import
// @Accept json
// @Produce json
// @Param body body service.ImportBatch true "Import batch"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /api/import/validate [post]
func (h *ImportHandler) Validate(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	batch, err := h.importer.ParseJSON(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	errors := h.importer.ValidateImport(*batch)

	c.JSON(http.StatusOK, gin.H{
		"valid":       len(errors) == 0,
		"errors":      errors,
		"errorCount":  len(errors),
		"totalCount":  len(batch.Articles),
	})
}

// GetTemplate godoc
// @Summary Get import template
// @Description Get a JSON template for article import
// @Tags import
// @Produce json
// @Success 200 {object} service.ImportBatch
// @Router /api/import/template [get]
func (h *ImportHandler) GetTemplate(c *gin.Context) {
	template := h.importer.GenerateTemplate()
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=import-template.json")
	c.String(http.StatusOK, template)
}

// Export godoc
// @Summary Export articles to JSON
// @Description Export articles in import-compatible JSON format
// @Tags import
// @Produce json
// @Param categoryId query string false "Filter by category ID"
// @Param status query string false "Filter by status"
// @Success 200 {object} service.ImportBatch
// @Router /api/import/export [get]
func (h *ImportHandler) Export(c *gin.Context) {
	var categoryID *uuid.UUID
	if catID := c.Query("categoryId"); catID != "" {
		if parsed, err := uuid.Parse(catID); err == nil {
			categoryID = &parsed
		}
	}

	status := c.Query("status")

	data, err := h.importer.BatchExportToJSON(categoryID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=articles-export.json")
	c.Data(http.StatusOK, "application/json", data)
}

// UploadFile godoc
// @Summary Upload and import JSON file
// @Description Upload a JSON file containing articles to import
// @Tags import
// @Accept multipart/form-data
// @Produce json
// @Param file formance file true "JSON file to import"
// @Param skipDuplicates formData bool false "Skip duplicate articles"
// @Param updateExisting formData bool false "Update existing articles"
// @Success 200 {object} service.ImportResult
// @Failure 400 {object} map[string]string
// @Router /api/import/upload [post]
func (h *ImportHandler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	// Check file extension
	if file.Filename[len(file.Filename)-5:] != ".json" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only JSON files are supported"})
		return
	}

	// Limit file size to 10MB
	if file.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file size exceeds 10MB limit"})
		return
	}

	// Open and read file
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	batch, err := h.importer.ParseJSON(data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Apply form options
	if c.PostForm("skipDuplicates") == "true" {
		batch.Options.SkipDuplicates = true
	}
	if c.PostForm("updateExisting") == "true" {
		batch.Options.UpdateExisting = true
	}

	result, err := h.importer.Import(*batch)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
