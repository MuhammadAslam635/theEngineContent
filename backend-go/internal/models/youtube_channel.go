package models

import (
	"encoding/json"
	"time"
)

// YouTubeChannel — full channel profile data from SociaVault.
type YouTubeChannel struct {
	ID                   int32           `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID               int32           `gorm:"not null" json:"user_id"`
	CompetitorAccountID  *int32          `json:"competitor_account_id,omitempty"`

	// Core identifiers
	ChannelID            string          `gorm:"size:255;uniqueIndex;not null" json:"channel_id"`
	ChannelURL           *string         `json:"channel_url,omitempty"`
	Handle               *string         `gorm:"size:255" json:"handle,omitempty"`
	Name                 string          `gorm:"size:255;not null" json:"name"`

	// Profile
	AvatarURL            *string         `json:"avatar_url,omitempty"`
	Description          *string         `json:"description,omitempty"`
	Country              *string         `gorm:"size:100" json:"country,omitempty"`
	JoinedDate           *string         `gorm:"size:100" json:"joined_date,omitempty"`

	// Stats
	SubscriberCount      int64           `gorm:"default:0" json:"subscriber_count"`
	SubscriberCountText  *string         `gorm:"size:50" json:"subscriber_count_text,omitempty"`
	VideoCount           int             `gorm:"default:0" json:"video_count"`
	ViewCount            int64           `gorm:"default:0" json:"view_count"`
	ViewCountText        *string         `gorm:"size:100" json:"view_count_text,omitempty"`

	// Metadata
	Tags                 *string         `json:"tags,omitempty"`
	Email                *string         `gorm:"size:255" json:"email,omitempty"`
	Links                json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"links"`

	// Tracking
	LastFetchedAt        time.Time       `gorm:"default:now()" json:"last_fetched_at"`
	IsActive             bool            `gorm:"default:true" json:"is_active"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`

	// Associations
	User              User               `gorm:"foreignKey:UserID" json:"-"`
	CompetitorAccount *CompetitorAccount  `gorm:"foreignKey:CompetitorAccountID" json:"competitor_account,omitempty"`
}
