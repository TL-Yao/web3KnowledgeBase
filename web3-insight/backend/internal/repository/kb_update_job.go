package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/model"
	"gorm.io/gorm"
)

type KBUpdateJobRepository struct {
	db *gorm.DB
}

func NewKBUpdateJobRepository(db *gorm.DB) *KBUpdateJobRepository {
	return &KBUpdateJobRepository{db: db}
}

// Create 创建新任务
func (r *KBUpdateJobRepository) Create(job *model.KBUpdateJob) error {
	return r.db.Create(job).Error
}

// GetByID 根据ID获取任务
func (r *KBUpdateJobRepository) GetByID(id uuid.UUID) (*model.KBUpdateJob, error) {
	var job model.KBUpdateJob
	err := r.db.Where("id = ?", id).First(&job).Error
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// Update 更新任务
func (r *KBUpdateJobRepository) Update(job *model.KBUpdateJob) error {
	return r.db.Save(job).Error
}

// UpdateStatus 更新任务状态和错误信息
func (r *KBUpdateJobRepository) UpdateStatus(id uuid.UUID, status string, errorMsg string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	// 根据状态设置时间字段
	now := time.Now()
	switch status {
	case "running":
		updates["started_at"] = now
	case "completed", "failed":
		updates["completed_at"] = now
	}

	if errorMsg != "" {
		updates["error_message"] = errorMsg
	}

	return r.db.
		Model(&model.KBUpdateJob{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// List 分页获取任务列表
func (r *KBUpdateJobRepository) List(page, pageSize int) ([]model.KBUpdateJob, int64, error) {
	var jobs []model.KBUpdateJob
	var total int64

	if err := r.db.Model(&model.KBUpdateJob{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := r.db.
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&jobs).Error

	return jobs, total, err
}

// GetLatest 获取最新的任务
func (r *KBUpdateJobRepository) GetLatest(limit int) ([]model.KBUpdateJob, error) {
	var jobs []model.KBUpdateJob
	err := r.db.
		Order("created_at DESC").
		Limit(limit).
		Find(&jobs).Error
	return jobs, err
}

// GetRunningJobs 获取正在运行的任务
func (r *KBUpdateJobRepository) GetRunningJobs() ([]model.KBUpdateJob, error) {
	var jobs []model.KBUpdateJob
	err := r.db.
		Where("status = ?", "running").
		Order("created_at DESC").
		Find(&jobs).Error
	return jobs, err
}

// UpdateProgress 更新任务进度
func (r *KBUpdateJobRepository) UpdateProgress(
	id uuid.UUID,
	keywordsGenerated int,
	articlesGenerated int,
	articlesPublished int,
) error {
	return r.db.
		Model(&model.KBUpdateJob{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"keywords_generated":  keywordsGenerated,
			"articles_generated":  articlesGenerated,
			"articles_published":  articlesPublished,
		}).Error
}

// AddSessionID 添加 session ID 到任务
func (r *KBUpdateJobRepository) AddSessionID(id uuid.UUID, sessionID string) error {
	return r.db.Exec(
		"UPDATE kb_update_jobs SET session_ids = array_append(session_ids, ?) WHERE id = ?",
		sessionID, id,
	).Error
}

// CleanupOrphanedJobs marks jobs that have been running for too long as failed
// This prevents zombie jobs from accumulating in the database
func (r *KBUpdateJobRepository) CleanupOrphanedJobs(maxRuntime time.Duration) (int64, error) {
	now := time.Now()
	cutoffTime := now.Add(-maxRuntime)

	result := r.db.
		Model(&model.KBUpdateJob{}).
		Where("status = ?", "running").
		Where("started_at < ?", cutoffTime).
		Updates(map[string]interface{}{
			"status":        "failed",
			"error_message": "Job timeout - process exceeded maximum runtime",
			"completed_at":  now,
		})

	return result.RowsAffected, result.Error
}

// FindByStatus finds all jobs with a given status
func (r *KBUpdateJobRepository) FindByStatus(status string) ([]model.KBUpdateJob, error) {
	var jobs []model.KBUpdateJob
	err := r.db.
		Where("status = ?", status).
		Order("created_at DESC").
		Find(&jobs).Error
	return jobs, err
}
