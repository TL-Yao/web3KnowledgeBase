package repository

import (
	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/model"
	"gorm.io/gorm"
)

type ExplorerRepository struct {
	db *gorm.DB
}

func NewExplorerRepository(db *gorm.DB) *ExplorerRepository {
	return &ExplorerRepository{db: db}
}

// List returns all explorer research entries with optional filters
func (r *ExplorerRepository) List(chainName, status string, limit int) ([]model.ExplorerResearch, error) {
	var explorers []model.ExplorerResearch

	query := r.db.Model(&model.ExplorerResearch{})

	if chainName != "" {
		query = query.Where("chain_name = ?", chainName)
	}
	if status != "" {
		query = query.Where("research_status = ?", status)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Order("chain_name ASC, popularity_score DESC").Find(&explorers).Error
	return explorers, err
}

// GetByID returns a single explorer research entry
func (r *ExplorerRepository) GetByID(id uuid.UUID) (*model.ExplorerResearch, error) {
	var explorer model.ExplorerResearch
	if err := r.db.First(&explorer, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &explorer, nil
}

// GetByURL returns an explorer research entry by URL
func (r *ExplorerRepository) GetByURL(url string) (*model.ExplorerResearch, error) {
	var explorer model.ExplorerResearch
	if err := r.db.First(&explorer, "explorer_url = ?", url).Error; err != nil {
		return nil, err
	}
	return &explorer, nil
}

// Create creates a new explorer research entry
func (r *ExplorerRepository) Create(explorer *model.ExplorerResearch) error {
	return r.db.Create(explorer).Error
}

// Update updates an existing explorer research entry
func (r *ExplorerRepository) Update(explorer *model.ExplorerResearch) error {
	return r.db.Save(explorer).Error
}

// Delete deletes an explorer research entry
func (r *ExplorerRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.ExplorerResearch{}, "id = ?", id).Error
}

// GetChains returns all unique chain names
func (r *ExplorerRepository) GetChains() ([]string, error) {
	var chains []string
	err := r.db.Model(&model.ExplorerResearch{}).
		Distinct("chain_name").
		Order("chain_name ASC").
		Pluck("chain_name", &chains).Error
	return chains, err
}

// GetByChain returns all explorers for a specific chain
func (r *ExplorerRepository) GetByChain(chainName string) ([]model.ExplorerResearch, error) {
	var explorers []model.ExplorerResearch
	err := r.db.Where("chain_name = ?", chainName).
		Order("popularity_score DESC").
		Find(&explorers).Error
	return explorers, err
}

// UpdateStatus updates the research status of an explorer
func (r *ExplorerRepository) UpdateStatus(id uuid.UUID, status string) error {
	return r.db.Model(&model.ExplorerResearch{}).
		Where("id = ?", id).
		Update("research_status", status).Error
}

// UpdateFeatures updates the features JSON of an explorer
func (r *ExplorerRepository) UpdateFeatures(id uuid.UUID, features interface{}) error {
	return r.db.Model(&model.ExplorerResearch{}).
		Where("id = ?", id).
		Update("features", features).Error
}

// Search searches explorers by name or chain
func (r *ExplorerRepository) Search(query string, limit int) ([]model.ExplorerResearch, error) {
	var explorers []model.ExplorerResearch
	err := r.db.Where("chain_name ILIKE ? OR explorer_name ILIKE ?", "%"+query+"%", "%"+query+"%").
		Order("popularity_score DESC").
		Limit(limit).
		Find(&explorers).Error
	return explorers, err
}

// GetStats returns statistics about the explorer research
func (r *ExplorerRepository) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total count
	var total int64
	if err := r.db.Model(&model.ExplorerResearch{}).Count(&total).Error; err != nil {
		return nil, err
	}
	stats["total"] = total

	// Count by status
	var statusCounts []struct {
		Status string
		Count  int64
	}
	if err := r.db.Model(&model.ExplorerResearch{}).
		Select("research_status as status, count(*) as count").
		Group("research_status").
		Scan(&statusCounts).Error; err != nil {
		return nil, err
	}
	statusMap := make(map[string]int64)
	for _, s := range statusCounts {
		statusMap[s.Status] = s.Count
	}
	stats["byStatus"] = statusMap

	// Count by chain
	var chainCounts []struct {
		Chain string
		Count int64
	}
	if err := r.db.Model(&model.ExplorerResearch{}).
		Select("chain_name as chain, count(*) as count").
		Group("chain_name").
		Order("count DESC").
		Limit(10).
		Scan(&chainCounts).Error; err != nil {
		return nil, err
	}
	stats["byChain"] = chainCounts

	return stats, nil
}

// FeatureRepository handles explorer feature definitions
type FeatureRepository struct {
	db *gorm.DB
}

func NewFeatureRepository(db *gorm.DB) *FeatureRepository {
	return &FeatureRepository{db: db}
}

// List returns all features
func (r *FeatureRepository) List() ([]model.ExplorerFeature, error) {
	var features []model.ExplorerFeature
	err := r.db.Order("category ASC, sort_order ASC").Find(&features).Error
	return features, err
}

// GetByCategory returns features by category
func (r *FeatureRepository) GetByCategory(category string) ([]model.ExplorerFeature, error) {
	var features []model.ExplorerFeature
	err := r.db.Where("category = ?", category).
		Order("sort_order ASC").
		Find(&features).Error
	return features, err
}

// Create creates a new feature
func (r *FeatureRepository) Create(feature *model.ExplorerFeature) error {
	return r.db.Create(feature).Error
}

// SeedStandardFeatures seeds the database with standard features
func (r *FeatureRepository) SeedStandardFeatures() error {
	for _, f := range model.StandardFeatures {
		var existing model.ExplorerFeature
		if err := r.db.Where("name = ? AND category = ?", f.Name, f.Category).First(&existing).Error; err == gorm.ErrRecordNotFound {
			if err := r.db.Create(&f).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
