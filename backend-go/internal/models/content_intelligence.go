package models

import (
	"time"

	"github.com/lib/pq"
)

// ---------------------------------------------------------------------------
// HookLibrary — data-driven hooks extracted from approved outlier reels
// ---------------------------------------------------------------------------

type HookLibrary struct {
	ID               int32          `gorm:"primaryKey;autoIncrement" json:"id"`
	OutlierReelID    int32          `gorm:"not null" json:"outlier_reel_id"`
	HookText         string         `gorm:"not null" json:"hook_text"`
	VisualHook       *string        `json:"visual_hook,omitempty"`
	TimingSeconds    *float64       `gorm:"type:numeric(5,2)" json:"timing_seconds,omitempty"`
	Pacing           *string        `gorm:"size:50" json:"pacing,omitempty"` // fast | medium | slow
	EmotionalTrigger *string        `gorm:"size:50" json:"emotional_trigger,omitempty"`
	TopicTags        pq.StringArray `gorm:"type:text[]" json:"topic_tags"`
	UsageCount       int            `gorm:"default:0" json:"usage_count"`
	AvgPerformance   float64        `gorm:"type:numeric(10,2);default:0" json:"avg_performance"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`

	// Associations
	OutlierReel OutlierReel `gorm:"foreignKey:OutlierReelID" json:"outlier_reel,omitempty"`
}

// EmotionalTrigger constants
const (
	EmotionalTriggerFrustration = "frustration"
	EmotionalTriggerCuriosity   = "curiosity"
	EmotionalTriggerFear        = "fear"
	EmotionalTriggerAspiration  = "aspiration"
)

// Pacing constants
const (
	PacingFast   = "fast"
	PacingMedium = "medium"
	PacingSlow   = "slow"
)

// ---------------------------------------------------------------------------
// AngleLibrary — data-driven angles, seeded with 11 base classifications
// ---------------------------------------------------------------------------

type AngleLibrary struct {
	ID                      int32      `gorm:"primaryKey;autoIncrement" json:"id"`
	OutlierReelID           *int32     `json:"outlier_reel_id,omitempty"` // NULL for seeded base angles
	AngleKey                string     `gorm:"uniqueIndex;size:100;not null" json:"angle_key"`
	DisplayName             string     `gorm:"size:150;not null" json:"display_name"`
	PsychologicalMechanism  string     `gorm:"not null" json:"psychological_mechanism"`
	AvoidWhen               *string    `json:"avoid_when,omitempty"`
	UsageCount              int        `gorm:"default:0" json:"usage_count"`
	AvgPerformance          float64    `gorm:"type:numeric(10,2);default:0" json:"avg_performance"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`

	// Associations
	OutlierReel *OutlierReel `gorm:"foreignKey:OutlierReelID" json:"outlier_reel,omitempty"`
}

// AngleKey constants — closed list, agents select only from these
const (
	AngleKeyContrarianArgument    = "contrarian_argument"
	AngleKeyHardTruth             = "hard_truth"
	AngleKeyFrameworkReveal       = "framework_reveal"
	AngleKeyVulnerabilityConfession = "vulnerability_confession"
	AngleKeyCaseStudy             = "case_study"
	AngleKeyMythBusting           = "myth_busting"
	AngleKeyPrediction            = "prediction"
	AngleKeyDirectChallenge       = "direct_challenge"
	AngleKeyBehindTheScenes       = "behind_the_scenes"
	AngleKeyReactionCommentary    = "reaction_commentary"
	AngleKeySpecificNumberProof   = "specific_number_proof"
)

// ---------------------------------------------------------------------------
// Persona — four target personas, agents select from this closed list
// ---------------------------------------------------------------------------

type Persona struct {
	ID           int32     `gorm:"primaryKey;autoIncrement" json:"id"`
	PersonaKey   string    `gorm:"uniqueIndex;size:100;not null" json:"persona_key"`
	DisplayName  string    `gorm:"size:150;not null" json:"display_name"`
	WhoTheyAre   string    `gorm:"not null" json:"who_they_are"`
	CorePain     string    `gorm:"not null" json:"core_pain"`
	IncomeRange  *string   `gorm:"size:100" json:"income_range,omitempty"`
	LanguageNotes *string  `json:"language_notes,omitempty"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// PersonaKey constants — closed list
const (
	PersonaKeyW2Escapee           = "w2_escapee"
	PersonaKeyStuckInvestor       = "stuck_investor"
	PersonaKeyAspiringEntrepreneur = "aspiring_entrepreneur"
	PersonaKeyIndustrySwitcher    = "industry_switcher"
)

// ---------------------------------------------------------------------------
// NickCredential — verified facts agents may use in scripts; never fabricated
// ---------------------------------------------------------------------------

type NickCredential struct {
	ID             int32      `gorm:"primaryKey;autoIncrement" json:"id"`
	CredentialKey  string     `gorm:"uniqueIndex;size:150;not null" json:"credential_key"`
	Category       string     `gorm:"size:100;not null" json:"category"` // results | experience | social_proof | financial
	DisplayValue   string     `gorm:"not null" json:"display_value"`
	RawValue       *float64   `gorm:"type:numeric" json:"raw_value,omitempty"`
	Unit           *string    `gorm:"size:50" json:"unit,omitempty"`
	Context        *string    `json:"context,omitempty"`
	VerifiedAt     time.Time  `gorm:"default:now()" json:"verified_at"`
	IsActive       bool       `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
