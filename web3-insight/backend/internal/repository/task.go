package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/user/web3-insight/internal/model"
	"gorm.io/gorm"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

type TaskListParams struct {
	Type   string
	Status string
	Page   int
	Limit  int
}

type TaskListResult struct {
	Tasks    []model.Task `json:"tasks"`
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"pageSize"`
}

func (r *TaskRepository) List(params TaskListParams) (*TaskListResult, error) {
	query := r.db.Model(&model.Task{})

	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}

	offset := (params.Page - 1) * params.Limit

	var tasks []model.Task
	if err := query.Order("created_at DESC").Offset(offset).Limit(params.Limit).Find(&tasks).Error; err != nil {
		return nil, err
	}

	return &TaskListResult{
		Tasks:    tasks,
		Total:    total,
		Page:     params.Page,
		PageSize: params.Limit,
	}, nil
}

func (r *TaskRepository) GetByID(id uuid.UUID) (*model.Task, error) {
	var task model.Task
	if err := r.db.First(&task, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *TaskRepository) Create(task *model.Task) error {
	return r.db.Create(task).Error
}

func (r *TaskRepository) Update(task *model.Task) error {
	return r.db.Save(task).Error
}

func (r *TaskRepository) Cancel(id uuid.UUID) error {
	return r.db.Model(&model.Task{}).Where("id = ? AND status IN ?", id, []string{model.TaskStatusPending, model.TaskStatusRunning}).Update("status", "cancelled").Error
}

type TaskStats struct {
	TotalTasks     int64           `json:"totalTasks"`
	PendingTasks   int64           `json:"pendingTasks"`
	RunningTasks   int64           `json:"runningTasks"`
	CompletedTasks int64           `json:"completedTasks"`
	FailedTasks    int64           `json:"failedTasks"`
	TotalCostUSD   decimal.Decimal `json:"totalCostUsd"`
	TotalTokens    int64           `json:"totalTokens"`
	TasksByType    map[string]int  `json:"tasksByType"`
}

func (r *TaskRepository) GetStats() (*TaskStats, error) {
	var stats TaskStats

	// Total tasks
	if err := r.db.Model(&model.Task{}).Count(&stats.TotalTasks).Error; err != nil {
		return nil, err
	}

	// Tasks by status - use single aggregated query for efficiency
	type statusCount struct {
		Status string
		Count  int64
	}
	var statusCounts []statusCount
	if err := r.db.Model(&model.Task{}).Select("status, COUNT(*) as count").Group("status").Scan(&statusCounts).Error; err != nil {
		return nil, err
	}
	for _, sc := range statusCounts {
		switch sc.Status {
		case model.TaskStatusPending:
			stats.PendingTasks = sc.Count
		case model.TaskStatusRunning:
			stats.RunningTasks = sc.Count
		case model.TaskStatusCompleted:
			stats.CompletedTasks = sc.Count
		case model.TaskStatusFailed:
			stats.FailedTasks = sc.Count
		}
	}

	// Total cost and tokens
	var costResult struct {
		TotalCost   decimal.Decimal
		TotalTokens int64
	}
	if err := r.db.Model(&model.Task{}).Select("COALESCE(SUM(cost_usd), 0) as total_cost, COALESCE(SUM(tokens_used), 0) as total_tokens").Scan(&costResult).Error; err != nil {
		return nil, err
	}
	stats.TotalCostUSD = costResult.TotalCost
	stats.TotalTokens = costResult.TotalTokens

	// Tasks by type
	var typeCounts []struct {
		Type  string
		Count int
	}
	if err := r.db.Model(&model.Task{}).Select("type, COUNT(*) as count").Group("type").Scan(&typeCounts).Error; err != nil {
		return nil, err
	}
	stats.TasksByType = make(map[string]int)
	for _, tc := range typeCounts {
		stats.TasksByType[tc.Type] = tc.Count
	}

	return &stats, nil
}

func (r *TaskRepository) CleanupOldTasks(olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	return r.db.Where("status IN ? AND created_at < ?", []string{model.TaskStatusCompleted, model.TaskStatusFailed}, cutoff).Delete(&model.Task{}).Error
}
