package repository

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/model"
	"gorm.io/gorm"
)

type ResearchSessionRepository struct {
	db *gorm.DB
}

func NewResearchSessionRepository(db *gorm.DB) *ResearchSessionRepository {
	return &ResearchSessionRepository{db: db}
}

func (r *ResearchSessionRepository) Create(session *model.ResearchSession) error {
	return r.db.Create(session).Error
}

func (r *ResearchSessionRepository) GetByID(id uuid.UUID) (*model.ResearchSession, error) {
	var session model.ResearchSession
	if err := r.db.First(&session, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *ResearchSessionRepository) Update(session *model.ResearchSession) error {
	return r.db.Save(session).Error
}

// UpdateStatus performs a targeted update of status/stage/stageDetail columns only,
// avoiding race conditions during generation.
func (r *ResearchSessionRepository) UpdateStatus(id uuid.UUID, status, stage, stageDetail string) error {
	return r.db.Model(&model.ResearchSession{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       status,
			"stage":        stage,
			"stage_detail": stageDetail,
		}).Error
}

// SetArticleID links a completed session to its generated article.
func (r *ResearchSessionRepository) SetArticleID(id uuid.UUID, articleID uuid.UUID) error {
	return r.db.Model(&model.ResearchSession{}).
		Where("id = ?", id).
		Update("article_id", articleID).Error
}

// UpdatePinnedFindings replaces the pinned findings JSONB field.
func (r *ResearchSessionRepository) UpdatePinnedFindings(id uuid.UUID, findings json.RawMessage) error {
	return r.db.Model(&model.ResearchSession{}).
		Where("id = ?", id).
		Update("pinned_findings", findings).Error
}

// List returns paginated sessions ordered by created_at DESC.
func (r *ResearchSessionRepository) List(page, pageSize int) ([]model.ResearchSession, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	var total int64
	if err := r.db.Model(&model.ResearchSession{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize

	var sessions []model.ResearchSession
	if err := r.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&sessions).Error; err != nil {
		return nil, 0, err
	}

	return sessions, total, nil
}

func (r *ResearchSessionRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.ResearchSession{}, "id = ?", id).Error
}

// CountActive returns the number of sessions in active states (planning/researching/writing).
func (r *ResearchSessionRepository) CountActive() (int64, error) {
	var count int64
	err := r.db.Model(&model.ResearchSession{}).
		Where("status IN ?", []string{"planning", "researching", "writing"}).
		Count(&count).Error
	return count, err
}

// CleanupOrphaned marks sessions stuck in active status for more than the given
// duration as "failed". Returns the number of cleaned up sessions.
func (r *ResearchSessionRepository) CleanupOrphaned(maxAge time.Duration) (int64, error) {
	cutoff := time.Now().Add(-maxAge)
	result := r.db.Model(&model.ResearchSession{}).
		Where("status IN ? AND updated_at < ?", []string{"planning", "researching", "writing"}, cutoff).
		Updates(map[string]interface{}{
			"status":        "failed",
			"error_message": "Session timed out (server restart)",
		})
	return result.RowsAffected, result.Error
}
