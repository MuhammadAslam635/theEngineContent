package models

import "time"

type CompetitorAccount struct {
	ID               int32      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID           int32      `gorm:"not null" json:"user_id"`
	PlatformID       int32      `gorm:"not null" json:"platform_id"`
	Handle           string     `gorm:"size:255;not null" json:"handle"`
	DisplayName      string     `gorm:"size:255;not null" json:"display_name"`
	ProfileURL       *string    `json:"profile_url,omitempty"`
	Niche            string     `gorm:"size:100;default:business_education" json:"niche"`
	AvgViewCount     int64      `gorm:"default:0" json:"avg_view_count"`
	OutlierThreshold int64      `gorm:"default:0" json:"outlier_threshold"`
	LastScannedAt    *time.Time `json:"last_scanned_at,omitempty"`
	IsActive         bool       `gorm:"default:true" json:"is_active"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	// Associations
	Platform     Platform      `gorm:"foreignKey:PlatformID" json:"platform,omitempty"`
	OutlierReels []OutlierReel `gorm:"foreignKey:CompetitorAccountID" json:"-"`
}
