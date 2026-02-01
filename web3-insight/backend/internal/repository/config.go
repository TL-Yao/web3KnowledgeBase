package repository

import (
	"encoding/json"

	"github.com/user/web3-insight/internal/model"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ConfigRepository struct {
	db *gorm.DB
}

func NewConfigRepository(db *gorm.DB) *ConfigRepository {
	return &ConfigRepository{db: db}
}

func (r *ConfigRepository) GetAll() ([]model.Config, error) {
	var configs []model.Config
	if err := r.db.Order("key ASC").Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

func (r *ConfigRepository) Get(key string) (*model.Config, error) {
	var config model.Config
	if err := r.db.First(&config, "key = ?", key).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *ConfigRepository) Set(key, value, description string) error {
	var config model.Config
	result := r.db.First(&config, "key = ?", key)

	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if result.Error == gorm.ErrRecordNotFound {
		config = model.Config{
			Key:         key,
			Value:       datatypes.JSON(jsonValue),
			Description: description,
		}
		return r.db.Create(&config).Error
	}

	if result.Error != nil {
		return result.Error
	}

	config.Value = datatypes.JSON(jsonValue)
	if description != "" {
		config.Description = description
	}
	return r.db.Save(&config).Error
}

func (r *ConfigRepository) Delete(key string) error {
	return r.db.Delete(&model.Config{}, "key = ?", key).Error
}

func (r *ConfigRepository) GetMap() (map[string]string, error) {
	configs, err := r.GetAll()
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, c := range configs {
		var value string
		if err := json.Unmarshal(c.Value, &value); err != nil {
			// If it's not a simple string, use the raw JSON
			result[c.Key] = string(c.Value)
		} else {
			result[c.Key] = value
		}
	}
	return result, nil
}

func (r *ConfigRepository) SetMultiple(configs map[string]string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for key, value := range configs {
			var config model.Config
			result := tx.First(&config, "key = ?", key)

			jsonValue, err := json.Marshal(value)
			if err != nil {
				return err
			}

			if result.Error == gorm.ErrRecordNotFound {
				config = model.Config{
					Key:   key,
					Value: datatypes.JSON(jsonValue),
				}
				if err := tx.Create(&config).Error; err != nil {
					return err
				}
			} else if result.Error != nil {
				return result.Error
			} else {
				config.Value = datatypes.JSON(jsonValue)
				if err := tx.Save(&config).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
