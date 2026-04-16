package models

import (
	"time"

	"gorm.io/gorm"
)

type AiTask struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int32          `gorm:"not null" json:"user_id"`
	TaskType  string         `gorm:"size:50;not null" json:"task_type"`
	Status    string         `gorm:"size:20;default:pending" json:"status"`
	Payload   string         `gorm:"type:jsonb" json:"payload"`
	Result    string         `gorm:"type:jsonb" json:"result"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
