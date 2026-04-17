package models

import (
	"time"

	"github.com/lib/pq"
)

// ---------------------------------------------------------------------------
// TrendSignal — 30d/7d/24h topic volume + stage classification
// ---------------------------------------------------------------------------

type TrendSignal struct {
	ID           int32      `gorm:"primaryKey;autoIncrement" json:"id"`
	Topic        string     `gorm:"size:255;not null" json:"topic"`
	PlatformID   int32      `gorm:"not null" json:"platform_id"`
	Volume30d    int        `gorm:"default:0" json:"volume_30d"`
	Volume7d     int        `gorm:"default:0" json:"volume_7d"`
	Volume24h    int        `gorm:"default:0" json:"volume_24h"`
	Stage        string     `gorm:"size:20;default:emerging" json:"stage"` // emerging | peaking | saturated | declining
	UrgencyRating int8     `gorm:"default:0" json:"urgency_rating"`
	DetectedAt   time.Time  `gorm:"default:now()" json:"detected_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// Associations
	Platform Platform `gorm:"foreignKey:PlatformID" json:"platform,omitempty"`
}

// TrendStage constants — closed list
const (
	TrendStageEmerging  = "emerging"
	TrendStagePeaking   = "peaking"
	TrendStageSaturated = "saturated"
	TrendStageDeclining = "declining"
)

// ---------------------------------------------------------------------------
// ApprovedScript — the storehouse of scripts always banked and ready to render
// ---------------------------------------------------------------------------

type ApprovedScript struct {
	ID                    int32          `gorm:"primaryKey;autoIncrement" json:"id"`
	OutlierReelID         *int32         `json:"outlier_reel_id,omitempty"`
	Title                 string         `gorm:"size:255;not null" json:"title"`
	ScriptBody            string         `gorm:"not null" json:"script_body"`
	ContentType           string         `gorm:"size:20;default:short_form" json:"content_type"` // short_form | long_form
	Source                string         `gorm:"size:30;default:outlier_copy" json:"source"`     // outlier_copy | writer_agent | manual
	PersonaID             *int32         `json:"persona_id,omitempty"`
	AngleID               *int32         `json:"angle_id,omitempty"`
	HookID                *int32         `json:"hook_id,omitempty"`
	TopicTags             pq.StringArray `gorm:"type:text[]" json:"topic_tags"`
	WordCount             int            `gorm:"default:0" json:"word_count"`
	EstimatedDurationSecs int            `gorm:"default:0" json:"estimated_duration_seconds"`
	VoiceDNAScore         float64        `gorm:"type:numeric(5,2);default:0" json:"voicedna_score"`
	ValidationScore       float64        `gorm:"type:numeric(5,2);default:0" json:"validation_score"`
	ValidationNotes       *string        `json:"validation_notes,omitempty"`
	TimesUsed             int            `gorm:"default:0" json:"times_used"`
	LastUsedAt            *time.Time     `json:"last_used_at,omitempty"`
	AvgViewsWhenUsed      float64        `gorm:"type:numeric(10,2);default:0" json:"avg_views_when_used"`
	IsActive              bool           `gorm:"default:true" json:"is_active"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`

	// Associations
	OutlierReel *OutlierReel  `gorm:"foreignKey:OutlierReelID" json:"outlier_reel,omitempty"`
	Persona     *Persona      `gorm:"foreignKey:PersonaID" json:"persona,omitempty"`
	Angle       *AngleLibrary `gorm:"foreignKey:AngleID" json:"angle,omitempty"`
	Hook        *HookLibrary  `gorm:"foreignKey:HookID" json:"hook,omitempty"`
}

// ScriptSource constants — closed list
const (
	ScriptSourceOutlierCopy  = "outlier_copy"
	ScriptSourceWriterAgent  = "writer_agent"
	ScriptSourceManual       = "manual"
)

// ContentType constants — reused across models
const (
	ContentTypeShortForm = "short_form"
	ContentTypeLongForm  = "long_form"
)

// ---------------------------------------------------------------------------
// ContentBrief — Agent 1 + Agent 2 outputs, one per approved idea
// ---------------------------------------------------------------------------

type ContentBrief struct {
	ID                   int32      `gorm:"primaryKey;autoIncrement" json:"id"`
	IdeaSource           *string    `gorm:"size:255" json:"idea_source,omitempty"`
	IdeaEngineRef        *string    `gorm:"size:255" json:"idea_engine_ref,omitempty"`

	// Agent 1 — Strategic Decisions
	PrimaryPlatformID    int32      `gorm:"not null" json:"primary_platform_id"`
	ContentType          string     `gorm:"size:20;default:short_form" json:"content_type"`
	PersonaID            int32      `gorm:"not null" json:"persona_id"`
	VideoType            string     `gorm:"size:20;default:avatar" json:"video_type"` // avatar | faceless
	TrendSignalID        *int32     `json:"trend_signal_id,omitempty"`
	OutlierReelRefID     *int32     `json:"outlier_reel_ref_id,omitempty"`
	Agent1Confidence     float64    `gorm:"type:numeric(5,2);default:0" json:"agent1_confidence"`
	Agent1Reasoning      *string    `json:"agent1_reasoning,omitempty"`

	// Agent 2 — Creative Decisions
	PrimaryAngleID       *int32     `json:"primary_angle_id,omitempty"`
	SecondaryAngleID     *int32     `json:"secondary_angle_id,omitempty"`
	PrimaryHookID        *int32     `json:"primary_hook_id,omitempty"`
	Agent2Confidence     float64    `gorm:"type:numeric(5,2);default:0" json:"agent2_confidence"`
	Agent2Reasoning      *string    `json:"agent2_reasoning,omitempty"`
	AvoidWhenVerified    bool       `gorm:"default:false" json:"avoid_when_verified"`

	// Script selection outcome
	SelectedScriptID     *int32     `json:"selected_script_id,omitempty"`
	ScriptSelectionMethod string    `gorm:"size:30;default:banked" json:"script_selection_method"` // banked | writer_agent
	ScriptAttempts       int8       `gorm:"default:0" json:"script_attempts"`
	FinalScriptScore     float64    `gorm:"type:numeric(5,2);default:0" json:"final_script_score"`

	Status               string     `gorm:"size:30;default:pending" json:"status"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`

	// Associations
	PrimaryPlatform  Platform        `gorm:"foreignKey:PrimaryPlatformID" json:"primary_platform,omitempty"`
	Persona          Persona         `gorm:"foreignKey:PersonaID" json:"persona,omitempty"`
	TrendSignal      *TrendSignal    `gorm:"foreignKey:TrendSignalID" json:"trend_signal,omitempty"`
	OutlierReelRef   *OutlierReel    `gorm:"foreignKey:OutlierReelRefID" json:"outlier_reel_ref,omitempty"`
	PrimaryAngle     *AngleLibrary   `gorm:"foreignKey:PrimaryAngleID" json:"primary_angle,omitempty"`
	SecondaryAngle   *AngleLibrary   `gorm:"foreignKey:SecondaryAngleID" json:"secondary_angle,omitempty"`
	PrimaryHook      *HookLibrary    `gorm:"foreignKey:PrimaryHookID" json:"primary_hook,omitempty"`
	SelectedScript   *ApprovedScript `gorm:"foreignKey:SelectedScriptID" json:"selected_script,omitempty"`
}

// ContentBrief status constants — closed list
const (
	BriefStatusPending         = "pending"
	BriefStatusAgent1Complete  = "agent1_complete"
	BriefStatusAgent2Complete  = "agent2_complete"
	BriefStatusScriptSelected  = "script_selected"
	BriefStatusProduction      = "production"
	BriefStatusReview          = "review"
	BriefStatusPosted          = "posted"
	BriefStatusFailed          = "failed"
)

// VideoType constants
const (
	VideoTypeAvatar   = "avatar"
	VideoTypeFaceless = "faceless"
)

// ---------------------------------------------------------------------------
// ScriptGenerationAttempt — full audit trail per Writer Agent attempt
// ---------------------------------------------------------------------------

type ScriptGenerationAttempt struct {
	ID              int32          `gorm:"primaryKey;autoIncrement" json:"id"`
	ContentBriefID  int32          `gorm:"not null" json:"content_brief_id"`
	AttemptNumber   int8           `gorm:"not null" json:"attempt_number"`

	ScriptBody      string         `gorm:"not null" json:"script_body"`
	AdjustedScript  *string        `json:"adjusted_script,omitempty"`

	// Five-dimension scores
	ScoreIdeaAlignment float64     `gorm:"type:numeric(5,2);default:0" json:"score_idea_alignment"`
	ScoreAngleMatch    float64     `gorm:"type:numeric(5,2);default:0" json:"score_angle_match"`
	ScoreHookMatch     float64     `gorm:"type:numeric(5,2);default:0" json:"score_hook_match"`
	ScorePersonaFit    float64     `gorm:"type:numeric(5,2);default:0" json:"score_persona_fit"`
	ScoreVoiceDNA      float64     `gorm:"type:numeric(5,2);default:0" json:"score_voicedna"`
	OverallScore       float64     `gorm:"type:numeric(5,2);default:0" json:"overall_score"`

	IssuesFound     pq.StringArray `gorm:"type:text[]" json:"issues_found"`
	PassedThreshold bool           `gorm:"default:false" json:"passed_threshold"`
	Selected        bool           `gorm:"default:false" json:"selected"`

	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`

	// Associations
	ContentBrief ContentBrief `gorm:"foreignKey:ContentBriefID" json:"content_brief,omitempty"`
}

// ScriptPassThreshold — the 90% minimum from the spec
const ScriptPassThreshold = 90.0
