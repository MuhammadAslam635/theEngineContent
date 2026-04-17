package dto

// ---------------------------------------------------------------------------
// Shared pagination
// ---------------------------------------------------------------------------

type PaginationRequest struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

func NewPaginatedResponse(data interface{}, total int64, page, pageSize int) PaginatedResponse {
	totalPages := 0
	if pageSize > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	return PaginatedResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}

// ---------------------------------------------------------------------------
// Intelligence — competitor accounts
// ---------------------------------------------------------------------------

type AddCompetitorAccountRequest struct {
	PlatformID  int32  `json:"platform_id" binding:"required"`
	Handle      string `json:"handle" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
	ProfileURL  string `json:"profile_url"`
	Niche       string `json:"niche"`
}

// ---------------------------------------------------------------------------
// Intelligence — outlier reels
// ---------------------------------------------------------------------------

type IngestReelRequest struct {
	CompetitorAccountID int32  `json:"competitor_account_id" binding:"required"`
	PlatformID          int32  `json:"platform_id" binding:"required"`
	ExternalReelID      string `json:"external_reel_id" binding:"required"`
	Title               string `json:"title"`
	URL                 string `json:"url"`
	ViewCount           int64  `json:"view_count" binding:"required"`
	LikeCount           int64  `json:"like_count"`
	CommentCount        int64  `json:"comment_count"`
	ShareCount          int64  `json:"share_count"`
	RawTranscript       string `json:"raw_transcript"`
}

type RelevanceDecisionRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type UpdateBaselineRequest struct {
	AvgViewCount int64 `json:"avg_view_count" binding:"required"`
}

// ---------------------------------------------------------------------------
// Intelligence — trend signals
// ---------------------------------------------------------------------------

type UpsertTrendSignalRequest struct {
	Topic         string `json:"topic" binding:"required"`
	PlatformID    int32  `json:"platform_id" binding:"required"`
	Volume30d     int    `json:"volume_30d"`
	Volume7d      int    `json:"volume_7d"`
	Volume24h     int    `json:"volume_24h"`
	Stage         string `json:"stage" binding:"required"`
	UrgencyRating int8   `json:"urgency_rating"`
}

// ---------------------------------------------------------------------------
// Content library — credentials
// ---------------------------------------------------------------------------

type AddCredentialRequest struct {
	CredentialKey string   `json:"credential_key" binding:"required"`
	Category      string   `json:"category" binding:"required"`
	DisplayValue  string   `json:"display_value" binding:"required"`
	RawValue      *float64 `json:"raw_value"`
	Unit          string   `json:"unit"`
	Context       string   `json:"context"`
}

// ---------------------------------------------------------------------------
// Content library — approved scripts
// ---------------------------------------------------------------------------

type AddScriptRequest struct {
	OutlierReelID         *int32   `json:"outlier_reel_id"`
	Title                 string   `json:"title" binding:"required"`
	ScriptBody            string   `json:"script_body" binding:"required"`
	ContentType           string   `json:"content_type" binding:"required"`
	Source                string   `json:"source"`
	PersonaID             *int32   `json:"persona_id"`
	AngleID               *int32   `json:"angle_id"`
	HookID                *int32   `json:"hook_id"`
	TopicTags             []string `json:"topic_tags"`
	WordCount             int      `json:"word_count"`
	EstimatedDurationSecs int      `json:"estimated_duration_seconds"`
	VoiceDNAScore         float64  `json:"voicedna_score" binding:"required"`
	ValidationScore       float64  `json:"validation_score"`
	ValidationNotes       string   `json:"validation_notes"`
}

// ---------------------------------------------------------------------------
// Pipeline — content briefs
// ---------------------------------------------------------------------------

type CreateBriefRequest struct {
	IdeaSource    string `json:"idea_source"`
	IdeaEngineRef string `json:"idea_engine_ref"`
}

type Agent1DecisionRequest struct {
	PrimaryPlatformID int32   `json:"primary_platform_id" binding:"required"`
	ContentType       string  `json:"content_type" binding:"required"`
	VideoType         string  `json:"video_type" binding:"required"`
	PersonaID         int32   `json:"persona_id" binding:"required"`
	TrendSignalID     *int32  `json:"trend_signal_id"`
	OutlierReelRefID  *int32  `json:"outlier_reel_ref_id"`
	Confidence        float64 `json:"confidence" binding:"required"`
	Reasoning         string  `json:"reasoning"`
}

type Agent2DecisionRequest struct {
	PrimaryAngleID    *int32  `json:"primary_angle_id"`
	SecondaryAngleID  *int32  `json:"secondary_angle_id"`
	PrimaryHookID     *int32  `json:"primary_hook_id"`
	Confidence        float64 `json:"confidence" binding:"required"`
	Reasoning         string  `json:"reasoning"`
	AvoidWhenVerified bool    `json:"avoid_when_verified"`
}

type RecordAttemptRequest struct {
	AttemptNumber      int8     `json:"attempt_number" binding:"required"`
	ScriptBody         string   `json:"script_body" binding:"required"`
	AdjustedScript     string   `json:"adjusted_script"`
	ScoreIdeaAlignment float64  `json:"score_idea_alignment"`
	ScoreAngleMatch    float64  `json:"score_angle_match"`
	ScoreHookMatch     float64  `json:"score_hook_match"`
	ScorePersonaFit    float64  `json:"score_persona_fit"`
	ScoreVoiceDNA      float64  `json:"score_voicedna"`
	IssuesFound        []string `json:"issues_found"`
}

// ---------------------------------------------------------------------------
// Pipeline — SSML
// ---------------------------------------------------------------------------

type CreateSSMLRequest struct {
	ApprovedScriptID  int32  `json:"approved_script_id" binding:"required"`
	SSMLBody          string `json:"ssml_body" binding:"required"`
	Platform          string `json:"platform"`
	ElevenLabsVoiceID string `json:"elevenlabs_voice_id"`
}

type UpdateAudioStatusRequest struct {
	Status   string  `json:"status" binding:"required"`
	AudioURL *string `json:"audio_url"`
	JobID    *string `json:"job_id"`
}

type MarkAudioReadyRequest struct {
	AudioURL        string  `json:"audio_url" binding:"required"`
	DurationSeconds float64 `json:"duration_seconds"`
}

// ---------------------------------------------------------------------------
// Pipeline — confidence log
// ---------------------------------------------------------------------------

type LogConfidenceRequest struct {
	AuditLogID      *int32  `json:"audit_log_id"`
	AgentName       string  `json:"agent_name" binding:"required"`
	DecisionType    string  `json:"decision_type" binding:"required"`
	ConfidenceScore float64 `json:"confidence_score" binding:"required"`
}

// ---------------------------------------------------------------------------
// Production — videos
// ---------------------------------------------------------------------------

type CreateVideoRequest struct {
	ContentBriefID int32  `json:"content_brief_id" binding:"required"`
	SSMLScriptID   int32  `json:"ssml_script_id" binding:"required"`
	ProductionTool string `json:"production_tool" binding:"required"`
	VideoType      string `json:"video_type" binding:"required"`
	HeyGenAvatarID string `json:"heygen_avatar_id"`
}

type UpdateRenderStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// ---------------------------------------------------------------------------
// Production — scenes
// ---------------------------------------------------------------------------

type CreateSceneRequest struct {
	SceneNumber           int8    `json:"scene_number" binding:"required"`
	SceneScript           string  `json:"scene_script" binding:"required"`
	DurationSeconds       float64 `json:"duration_seconds"`
	WardrobeDescription   string  `json:"wardrobe_description"`
	BackgroundDescription string  `json:"background_description"`
	LogoPresent           *bool   `json:"logo_present"`
}

type CreateScenesRequest struct {
	Scenes []CreateSceneRequest `json:"scenes" binding:"required"`
}

type UpdateSceneCoherenceRequest struct {
	Passed        bool   `json:"passed"`
	FailureReason string `json:"failure_reason"`
}

type UpdateSceneRenderedRequest struct {
	SceneURL string `json:"scene_url" binding:"required"`
}

// ---------------------------------------------------------------------------
// Production — posting packages
// ---------------------------------------------------------------------------

type CreatePostingPackageRequest struct {
	PlatformID        int32    `json:"platform_id" binding:"required"`
	Caption           string   `json:"caption" binding:"required"`
	HookLine          string   `json:"hook_line"`
	Hashtags          []string `json:"hashtags"`
	FormattedVideoURL string   `json:"formatted_video_url"`
}

type ReviewDecisionRequest struct {
	Action          string `json:"action" binding:"required"` // approve | reject
	RejectionReason string `json:"rejection_reason"`
}

type MarkPostedRequest struct {
	PlatformPostID string `json:"platform_post_id" binding:"required"`
}

// ---------------------------------------------------------------------------
// Production — analytics
// ---------------------------------------------------------------------------

type RecordAnalyticsRequest struct {
	Window         string  `json:"window" binding:"required"`
	ViewCount      int64   `json:"view_count"`
	LikeCount      int64   `json:"like_count"`
	CommentCount   int64   `json:"comment_count"`
	ShareCount     int64   `json:"share_count"`
	SaveCount      int64   `json:"save_count"`
	SendsPerReach  float64 `json:"sends_per_reach"`
	CompletionRate float64 `json:"completion_rate"`
	EngagementRate float64 `json:"engagement_rate"`
}
