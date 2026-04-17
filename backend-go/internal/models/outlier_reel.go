package models

import (
	"time"

	"github.com/lib/pq"
)

type OutlierReel struct {
	ID                    int32          `gorm:"primaryKey;autoIncrement" json:"id"`
	CompetitorAccountID   int32          `gorm:"not null" json:"competitor_account_id"`
	PlatformID            int32          `gorm:"not null" json:"platform_id"`
	ExternalReelID        string         `gorm:"size:255;not null" json:"external_reel_id"`
	Title                 *string        `json:"title,omitempty"`
	URL                   *string        `json:"url,omitempty"`
	ViewCount             int64          `gorm:"default:0" json:"view_count"`
	LikeCount             int64          `gorm:"default:0" json:"like_count"`
	CommentCount          int64          `gorm:"default:0" json:"comment_count"`
	ShareCount            int64          `gorm:"default:0" json:"share_count"`
	AccountAvgAtCapture   int64          `gorm:"default:0" json:"account_avg_at_capture"`
	Multiplier            float64        `gorm:"type:numeric(6,2);default:0" json:"multiplier"`
	RawTranscript         *string        `json:"raw_transcript,omitempty"`
	PublishedAt           *time.Time     `json:"published_at,omitempty"`
	DetectedAt            time.Time      `gorm:"default:now()" json:"detected_at"`
	RelevanceStatus       string         `gorm:"size:20;default:pending" json:"relevance_status"`
	RelevanceReason       *string        `json:"relevance_reason,omitempty"`
	RelevanceCheckedAt    *time.Time     `json:"relevance_checked_at,omitempty"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`

	// Associations
	CompetitorAccount CompetitorAccount `gorm:"foreignKey:CompetitorAccountID" json:"competitor_account,omitempty"`
	Platform          Platform          `gorm:"foreignKey:PlatformID" json:"platform,omitempty"`
	HookLibrary       []HookLibrary     `gorm:"foreignKey:OutlierReelID" json:"-"`
	AngleLibrary      []AngleLibrary    `gorm:"foreignKey:OutlierReelID" json:"-"`
	ApprovedScripts   []ApprovedScript  `gorm:"foreignKey:OutlierReelID" json:"-"`
}

// RelevanceStatus constants — closed list, agents never invent new values
const (
	RelevanceStatusPending  = "pending"
	RelevanceStatusApproved = "approved"
	RelevanceStatusRejected = "rejected"
)

// ensure pq is used for TEXT[] fields in other models
var _ = pq.StringArray{}
