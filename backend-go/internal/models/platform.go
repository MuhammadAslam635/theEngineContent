package models

import "time"

type Platform struct {
	ID          int32     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"uniqueIndex;size:100;not null" json:"name"`
	DisplayName string    `gorm:"size:100;not null" json:"display_name"`
	ChannelURL  *string   `json:"channel_url,omitempty"`
	APIType     string    `gorm:"size:50;not null" json:"api_type"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Associations
	CompetitorAccounts []CompetitorAccount `gorm:"foreignKey:PlatformID" json:"-"`
	TrendSignals       []TrendSignal       `gorm:"foreignKey:PlatformID" json:"-"`
	PostingPackages    []PostingPackage    `gorm:"foreignKey:PlatformID" json:"-"`
	VideoAnalytics     []VideoAnalytic     `gorm:"foreignKey:PlatformID" json:"-"`
}
