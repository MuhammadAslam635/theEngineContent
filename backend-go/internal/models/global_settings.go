package models

import (
	"time"

	"gorm.io/gorm"
)

type GlobalSettings struct {
	ID        int32          `gorm:"primaryKey;autoIncrement" json:"id"`
	KeyName   string         `gorm:"uniqueIndex;not null;size:255" json:"key_name"`
	KeyValue  string         `gorm:"type:text" json:"key_value"`
	KeyType   string         `gorm:"size:50;default:'string'" json:"key_type"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}