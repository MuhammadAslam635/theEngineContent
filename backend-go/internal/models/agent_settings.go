package models

import (
	"time"

	"gorm.io/gorm"
)

type AgentSettings struct {
	ID          int32          `gorm:"primaryKey;autoIncrement" json:"id"`
	AgentName   string         `gorm:"not null;size:255;index" json:"agent_name"`
	Supervisor  string         `gorm:"size:255;index" json:"supervisor"`
	Provider    string         `gorm:"not null;size:100" json:"provider"`
	ModelName   string         `gorm:"not null;size:255" json:"model_name"`
	Prompt      string         `gorm:"type:text" json:"prompt"`
	Temperature float32        `gorm:"default:0.7" json:"temperature"`
	MaxTokens   int32          `gorm:"default:2048" json:"max_tokens"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
