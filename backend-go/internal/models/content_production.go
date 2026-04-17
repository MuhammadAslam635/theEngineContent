package models

import (
	"time"

	"github.com/lib/pq"
)

// ---------------------------------------------------------------------------
// SSMLScript — SSML-formatted script ready for ElevenLabs submission
// ---------------------------------------------------------------------------

type SSMLScript struct {
	ID                   int32      `gorm:"primaryKey;autoIncrement" json:"id"`
	ContentBriefID       int32      `gorm:"uniqueIndex;not null" json:"content_brief_id"`
	ApprovedScriptID     int32      `gorm:"not null" json:"approved_script_id"`
	SSMLBody             string     `gorm:"not null" json:"ssml_body"`
	Platform             *string    `gorm:"size:20" json:"platform,omitempty"`
	ElevenLabsVoiceID    *string    `gorm:"size:255" json:"elevenlabs_voice_id,omitempty"`
	ElevenLabsAudioURL   *string    `json:"elevenlabs_audio_url,omitempty"`
	ElevenLabsJobID      *string    `gorm:"size:255" json:"elevenlabs_job_id,omitempty"`
	AudioStatus          string     `gorm:"size:20;default:pending" json:"audio_status"` // pending | generating | ready | failed
	AudioDurationSeconds *float64   `gorm:"type:numeric(8,2)" json:"audio_duration_seconds,omitempty"`
	SubmittedAt          *time.Time `json:"submitted_at,omitempty"`
	AudioReadyAt         *time.Time `json:"audio_ready_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`

	// Associations
	ContentBrief   ContentBrief   `gorm:"foreignKey:ContentBriefID" json:"content_brief,omitempty"`
	ApprovedScript ApprovedScript `gorm:"foreignKey:ApprovedScriptID" json:"approved_script,omitempty"`
}

// AudioStatus constants
const (
	AudioStatusPending    = "pending"
	AudioStatusGenerating = "generating"
	AudioStatusReady      = "ready"
	AudioStatusFailed     = "failed"
)

// ---------------------------------------------------------------------------
// ContentVideo — master record per produced video
// ---------------------------------------------------------------------------

type ContentVideo struct {
	ID                   int32      `gorm:"primaryKey;autoIncrement" json:"id"`
	ContentBriefID       int32      `gorm:"uniqueIndex;not null" json:"content_brief_id"`
	SSMLScriptID         int32      `gorm:"not null" json:"ssml_script_id"`

	ProductionTool       string     `gorm:"size:30;default:heygen_avatar_shots" json:"production_tool"`
	// heygen_avatar_shots | heygen_avatar_v | kling_ai | elevenlabs_studio
	VideoType            string     `gorm:"size:20;default:avatar" json:"video_type"` // avatar | faceless
	HeyGenJobID          *string    `gorm:"size:255" json:"heygen_job_id,omitempty"`
	HeyGenAvatarID       *string    `gorm:"size:255" json:"heygen_avatar_id,omitempty"`

	RenderStatus         string     `gorm:"size:20;default:pending" json:"render_status"`
	// pending | rendering | coherence_check | stitching | ready | failed
	CoherenceCheckPassed *bool      `json:"coherence_check_passed,omitempty"`
	CoherenceCheckNotes  *string    `json:"coherence_check_notes,omitempty"`

	VideoURL             *string    `json:"video_url,omitempty"`
	VideoDurationSeconds *float64   `gorm:"type:numeric(8,2)" json:"video_duration_seconds,omitempty"`
	FileSizeBytes        *int64     `json:"file_size_bytes,omitempty"`
	Resolution           *string    `gorm:"size:20" json:"resolution,omitempty"`
	Format               string     `gorm:"size:10;default:mp4" json:"format"`

	RenderStartedAt      *time.Time `json:"render_started_at,omitempty"`
	RenderCompletedAt    *time.Time `json:"render_completed_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`

	// Associations
	ContentBrief ContentBrief `gorm:"foreignKey:ContentBriefID" json:"content_brief,omitempty"`
	SSMLScript   SSMLScript   `gorm:"foreignKey:SSMLScriptID" json:"ssml_script,omitempty"`
	VideoScenes  []VideoScene `gorm:"foreignKey:ContentVideoID" json:"-"`
}

// ProductionTool constants — closed list
const (
	ProductionToolHeyGenAvatarShots  = "heygen_avatar_shots"
	ProductionToolHeyGenAvatarV      = "heygen_avatar_v"
	ProductionToolKlingAI            = "kling_ai"
	ProductionToolElevenLabsStudio   = "elevenlabs_studio"
)

// RenderStatus constants
const (
	RenderStatusPending        = "pending"
	RenderStatusRendering      = "rendering"
	RenderStatusCoherenceCheck = "coherence_check"
	RenderStatusStitching      = "stitching"
	RenderStatusReady          = "ready"
	RenderStatusFailed         = "failed"
)

// ---------------------------------------------------------------------------
// VideoScene — individual 15-second HeyGen Avatar Shots scenes
// ---------------------------------------------------------------------------

type VideoScene struct {
	ID                     int32      `gorm:"primaryKey;autoIncrement" json:"id"`
	ContentVideoID         int32      `gorm:"not null" json:"content_video_id"`
	SceneNumber            int8       `gorm:"not null" json:"scene_number"`
	SceneScript            string     `gorm:"not null" json:"scene_script"`
	DurationSeconds        float64    `gorm:"type:numeric(5,2);default:15.00" json:"duration_seconds"`

	HeyGenSceneID          *string    `gorm:"size:255" json:"heygen_scene_id,omitempty"`
	WardrobeDescription    *string    `json:"wardrobe_description,omitempty"`
	BackgroundDescription  *string    `json:"background_description,omitempty"`
	LogoPresent            *bool      `json:"logo_present,omitempty"`
	SceneURL               *string    `json:"scene_url,omitempty"`

	CoherenceStatus        string     `gorm:"size:20;default:pending" json:"coherence_status"`
	// pending | passed | failed | re_rendering
	CoherenceFailureReason *string    `json:"coherence_failure_reason,omitempty"`

	RenderStatus           string     `gorm:"size:20;default:pending" json:"render_status"`
	RenderAttempts         int8       `gorm:"default:0" json:"render_attempts"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`

	// Associations
	ContentVideo ContentVideo `gorm:"foreignKey:ContentVideoID" json:"content_video,omitempty"`
}

// CoherenceStatus constants
const (
	CoherenceStatusPending     = "pending"
	CoherenceStatusPassed      = "passed"
	CoherenceStatusFailed      = "failed"
	CoherenceStatusReRendering = "re_rendering"
)

// ---------------------------------------------------------------------------
// PostingPackage — complete posting-ready bundle per video per platform
// ---------------------------------------------------------------------------

type PostingPackage struct {
	ID                   int32          `gorm:"primaryKey;autoIncrement" json:"id"`
	ContentVideoID       int32          `gorm:"uniqueIndex;not null" json:"content_video_id"`
	ContentBriefID       int32          `gorm:"not null" json:"content_brief_id"`
	TrackingID           string         `gorm:"uniqueIndex;size:64;not null" json:"tracking_id"`

	PlatformID           int32          `gorm:"not null" json:"platform_id"`
	FormattedVideoURL    *string        `json:"formatted_video_url,omitempty"`
	Caption              string         `gorm:"not null" json:"caption"`
	HookLine             *string        `json:"hook_line,omitempty"`
	Hashtags             pq.StringArray `gorm:"type:text[]" json:"hashtags"`
	RecommendedPostTime  *time.Time     `json:"recommended_post_time,omitempty"`

	ReviewStatus         string         `gorm:"size:20;default:pending_review" json:"review_status"`
	// pending_review | approved | rejected | posted
	ReviewedByUserID     *int32         `json:"reviewed_by_user_id,omitempty"`
	ReviewedAt           *time.Time     `json:"reviewed_at,omitempty"`
	RejectionReason      *string        `json:"rejection_reason,omitempty"`

	PostedAt             *time.Time     `json:"posted_at,omitempty"`
	PlatformPostID       *string        `gorm:"size:255" json:"platform_post_id,omitempty"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`

	// Associations
	ContentVideo     ContentVideo `gorm:"foreignKey:ContentVideoID" json:"content_video,omitempty"`
	ContentBrief     ContentBrief `gorm:"foreignKey:ContentBriefID" json:"content_brief,omitempty"`
	Platform         Platform     `gorm:"foreignKey:PlatformID" json:"platform,omitempty"`
	ReviewedByUser   *User        `gorm:"foreignKey:ReviewedByUserID" json:"reviewed_by_user,omitempty"`
	VideoAnalytics   []VideoAnalytic `gorm:"foreignKey:PostingPackageID" json:"-"`
}

// ReviewStatus constants
const (
	ReviewStatusPendingReview = "pending_review"
	ReviewStatusApproved      = "approved"
	ReviewStatusRejected      = "rejected"
	ReviewStatusPosted        = "posted"
)

// ---------------------------------------------------------------------------
// VideoAnalytic — performance data at 24h, 7d, 30d intervals
// ---------------------------------------------------------------------------

type VideoAnalytic struct {
	ID                   int32      `gorm:"primaryKey;autoIncrement" json:"id"`
	PostingPackageID     int32      `gorm:"not null" json:"posting_package_id"`
	PlatformID           int32      `gorm:"not null" json:"platform_id"`
	TrackingID           string     `gorm:"size:64;not null" json:"tracking_id"`

	Window               string     `gorm:"size:10;not null" json:"window"` // 24h | 7d | 30d
	CollectedAt          time.Time  `gorm:"default:now()" json:"collected_at"`

	ViewCount            int64      `gorm:"default:0" json:"view_count"`
	LikeCount            int64      `gorm:"default:0" json:"like_count"`
	CommentCount         int64      `gorm:"default:0" json:"comment_count"`
	ShareCount           int64      `gorm:"default:0" json:"share_count"`
	SaveCount            int64      `gorm:"default:0" json:"save_count"`
	SendsPerReach        float64    `gorm:"type:numeric(8,4);default:0" json:"sends_per_reach"`
	CompletionRate       float64    `gorm:"type:numeric(5,2);default:0" json:"completion_rate"`
	EngagementRate       float64    `gorm:"type:numeric(5,2);default:0" json:"engagement_rate"`

	BeatsBaseline        *bool      `json:"beats_baseline,omitempty"`
	BaselineMultiplier   *float64   `gorm:"type:numeric(6,2)" json:"baseline_multiplier,omitempty"`
	TriggersLongform     bool       `gorm:"default:false" json:"triggers_longform"`
	CreatedAt            time.Time  `json:"created_at"`

	// Associations
	PostingPackage PostingPackage `gorm:"foreignKey:PostingPackageID" json:"posting_package,omitempty"`
	Platform       Platform       `gorm:"foreignKey:PlatformID" json:"platform,omitempty"`
}

// AnalyticsWindow constants
const (
	AnalyticsWindow24h = "24h"
	AnalyticsWindow7d  = "7d"
	AnalyticsWindow30d = "30d"
)

// NickBaselineViews — Phase 1 success metric (38K average)
const NickBaselineViews = 38000

// LongFormTriggerThreshold — 50K views in 7d triggers long-form task in Phase 2
const LongFormTriggerThreshold = 50000

// ---------------------------------------------------------------------------
// AgentConfidenceLog — per-decision confidence + escalation audit trail
// ---------------------------------------------------------------------------

type AgentConfidenceLog struct {
	ID               int32      `gorm:"primaryKey;autoIncrement" json:"id"`
	AuditLogID       *int32     `json:"audit_log_id,omitempty"`
	ContentBriefID   *int32     `json:"content_brief_id,omitempty"`
	AgentName        string     `gorm:"size:100;not null" json:"agent_name"`
	DecisionType     string     `gorm:"size:100;not null" json:"decision_type"`
	ConfidenceScore  float64    `gorm:"type:numeric(5,2);not null" json:"confidence_score"`
	BelowThreshold   bool       `gorm:"->" json:"below_threshold"` // generated column, read-only in Go
	FlaggedForReview bool       `gorm:"default:false" json:"flagged_for_review"`
	AdversarialAgrees *bool     `json:"adversarial_agrees,omitempty"`
	Escalated        bool       `gorm:"default:false" json:"escalated"`
	EscalationReason *string    `json:"escalation_reason,omitempty"`
	RetryCount       int8       `gorm:"default:0" json:"retry_count"`
	Resolved         bool       `gorm:"default:false" json:"resolved"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	// Associations
	AuditLog     *AuditLog     `gorm:"foreignKey:AuditLogID" json:"audit_log,omitempty"`
	ContentBrief *ContentBrief `gorm:"foreignKey:ContentBriefID" json:"content_brief,omitempty"`
}

// AgentName constants — closed list matching the five agents + auditor
const (
	AgentNameStrategic         = "strategic_agent"
	AgentNameCreative          = "creative_agent"
	AgentNameValidation        = "validation_agent"
	AgentNameWriter            = "writer_agent"
	AgentNameSSML              = "ssml_agent"
	AgentNameAdversarialAuditor = "adversarial_auditor"
)

// ConfidenceFlagThreshold — outputs below 70 are flagged for human review
const ConfidenceFlagThreshold = 70.0

// MaxAgentRetries — after 2 failed retries escalate to PM
const MaxAgentRetries = 2
