package services

import (
	"errors"
	"fmt"

	"backend-go/internal/models"
	"backend-go/internal/repositories"
)

// ---------------------------------------------------------------------------
// PipelineService
// Orchestrates Sections 7 & 8: the script selection loop, content brief lifecycle,
// SSML formatting stage, and agent confidence gating.
//
// Engineering rule: nothing proceeds past a failed confidence gate.
// Hallucination is prevented through closed-list selection + adversarial auditing.
// ---------------------------------------------------------------------------

type PipelineService interface {
	// Content brief lifecycle
	CreateBrief(req CreateBriefRequest) (*models.ContentBrief, error)
	GetBrief(id int32) (*models.ContentBrief, error)
	ListBriefsByStatus(status string, page, pageSize int) ([]models.ContentBrief, int64, error)
	RecordAgent1Decision(briefID int32, req Agent1DecisionRequest) error
	RecordAgent2Decision(briefID int32, req Agent2DecisionRequest) error

	// Script selection loop (Section 8)
	RunScriptSelectionLoop(briefID int32) (*ScriptSelectionResult, error)
	RecordGenerationAttempt(req RecordAttemptRequest) (*models.ScriptGenerationAttempt, error)
	SelectBestAttempt(briefID int32) (*models.ScriptGenerationAttempt, error)

	// SSML stage
	CreateSSMLScript(req CreateSSMLRequest) (*models.SSMLScript, error)
	UpdateAudioStatus(ssmlID int32, status string, audioURL, jobID *string) error
	MarkAudioReady(ssmlID int32, audioURL string, durationSeconds float64) error

	// Agent confidence gating
	LogAgentConfidence(req LogConfidenceRequest) (*models.AgentConfidenceLog, error)
	ListFlaggedDecisions(resolved bool, page, pageSize int) ([]models.AgentConfidenceLog, int64, error)
	ListEscalatedDecisions(page, pageSize int) ([]models.AgentConfidenceLog, int64, error)
	ResolveConfidenceFlag(id int32) error
}

// ---------------------------------------------------------------------------
// Request / response types
// ---------------------------------------------------------------------------

type CreateBriefRequest struct {
	IdeaSource    string
	IdeaEngineRef string
}

type Agent1DecisionRequest struct {
	PrimaryPlatformID int32
	ContentType       string
	VideoType         string
	PersonaID         int32
	TrendSignalID     *int32
	OutlierReelRefID  *int32
	Confidence        float64
	Reasoning         string
}

type Agent2DecisionRequest struct {
	PrimaryAngleID    *int32
	SecondaryAngleID  *int32
	PrimaryHookID     *int32
	Confidence        float64
	Reasoning         string
	AvoidWhenVerified bool
}

type RecordAttemptRequest struct {
	ContentBriefID     int32
	AttemptNumber      int8
	ScriptBody         string
	AdjustedScript     string
	ScoreIdeaAlignment float64
	ScoreAngleMatch    float64
	ScoreHookMatch     float64
	ScorePersonaFit    float64
	ScoreVoiceDNA      float64
	IssuesFound        []string
}

type ScriptSelectionResult struct {
	SelectedScriptID     *int32  // set if a banked script was used
	SelectedAttemptID    *int32  // set if writer agent was used
	Method               string  // banked | writer_agent
	Attempts             int8
	FinalScore           float64
	PassedThreshold      bool
}

type CreateSSMLRequest struct {
	ContentBriefID    int32
	ApprovedScriptID  int32
	SSMLBody          string
	Platform          string
	ElevenLabsVoiceID string
}

type LogConfidenceRequest struct {
	AuditLogID      *int32
	ContentBriefID  *int32
	AgentName       string
	DecisionType    string
	ConfidenceScore float64
}

// ---------------------------------------------------------------------------
// Implementation
// ---------------------------------------------------------------------------

type pipelineService struct {
	briefRepo    repositories.ContentBriefRepository
	attemptRepo  repositories.ScriptGenerationAttemptRepository
	ssmlRepo     repositories.SSMLScriptRepository
	scriptRepo   repositories.ApprovedScriptRepository
	confidenceRepo repositories.AgentConfidenceLogRepository
}

func NewPipelineService(
	briefRepo repositories.ContentBriefRepository,
	attemptRepo repositories.ScriptGenerationAttemptRepository,
	ssmlRepo repositories.SSMLScriptRepository,
	scriptRepo repositories.ApprovedScriptRepository,
	confidenceRepo repositories.AgentConfidenceLogRepository,
) PipelineService {
	return &pipelineService{
		briefRepo:      briefRepo,
		attemptRepo:    attemptRepo,
		ssmlRepo:       ssmlRepo,
		scriptRepo:     scriptRepo,
		confidenceRepo: confidenceRepo,
	}
}

// --- Content brief lifecycle ---

func (s *pipelineService) CreateBrief(req CreateBriefRequest) (*models.ContentBrief, error) {
	var ideaSourcePtr, ideaEngineRefPtr *string
	if req.IdeaSource != "" {
		ideaSourcePtr = &req.IdeaSource
	}
	if req.IdeaEngineRef != "" {
		ideaEngineRefPtr = &req.IdeaEngineRef
	}
	brief := &models.ContentBrief{
		IdeaSource:    ideaSourcePtr,
		IdeaEngineRef: ideaEngineRefPtr,
		Status:        models.BriefStatusPending,
	}
	if err := s.briefRepo.Create(brief); err != nil {
		return nil, err
	}
	return brief, nil
}

func (s *pipelineService) GetBrief(id int32) (*models.ContentBrief, error) {
	return s.briefRepo.FindByID(id)
}

func (s *pipelineService) ListBriefsByStatus(status string, page, pageSize int) ([]models.ContentBrief, int64, error) {
	if err := validateBriefStatus(status); err != nil {
		return nil, 0, err
	}
	return s.briefRepo.FindByStatus(status, page, pageSize)
}

func (s *pipelineService) RecordAgent1Decision(briefID int32, req Agent1DecisionRequest) error {
	if err := validateVideoType(req.VideoType); err != nil {
		return err
	}
	if err := validateContentType(req.ContentType); err != nil {
		return err
	}
	return s.briefRepo.UpdateAgent1Output(
		briefID,
		req.PrimaryPlatformID,
		req.ContentType,
		req.VideoType,
		req.PersonaID,
		req.TrendSignalID,
		req.OutlierReelRefID,
		req.Confidence,
		req.Reasoning,
	)
}

func (s *pipelineService) RecordAgent2Decision(briefID int32, req Agent2DecisionRequest) error {
	return s.briefRepo.UpdateAgent2Output(
		briefID,
		req.PrimaryAngleID,
		req.SecondaryAngleID,
		req.PrimaryHookID,
		req.Confidence,
		req.Reasoning,
		req.AvoidWhenVerified,
	)
}

// --- Script selection loop (Section 8) ---

// RunScriptSelectionLoop queries the approved scripts database, scores candidates,
// and activates the Writer Agent path if no banked script reaches 90%.
// Returns the selection result for the orchestration layer to act on.
func (s *pipelineService) RunScriptSelectionLoop(briefID int32) (*ScriptSelectionResult, error) {
	brief, err := s.briefRepo.FindByID(briefID)
	if err != nil {
		return nil, fmt.Errorf("brief not found: %w", err)
	}

	// Query banked scripts matching the brief's content type, persona, and angle
	candidates, err := s.scriptRepo.FindBestForBrief(
		brief.ContentType,
		&brief.PersonaID,
		brief.PrimaryAngleID,
	)
	if err != nil {
		return nil, err
	}

	// Find the highest-scoring banked script
	bestScore, bestScriptID := findBestBankedScript(candidates)

	if bestScore >= models.ScriptPassThreshold {
		// Banked script passes — use it directly
		if err := s.briefRepo.UpdateScriptSelection(briefID, &bestScriptID, models.ScriptSourceOutlierCopy, 0, bestScore); err != nil {
			return nil, err
		}
		return &ScriptSelectionResult{
			SelectedScriptID: &bestScriptID,
			Method:           "banked",
			FinalScore:       bestScore,
			PassedThreshold:  true,
		}, nil
	}

	// No banked script reaches 90% — signal to orchestration that Writer Agent must activate.
	// Attempt recording and selection happen via RecordGenerationAttempt + SelectBestAttempt.
	return &ScriptSelectionResult{
		Method:          "writer_agent",
		FinalScore:      bestScore,
		PassedThreshold: false,
	}, nil
}

func (s *pipelineService) RecordGenerationAttempt(req RecordAttemptRequest) (*models.ScriptGenerationAttempt, error) {
	if req.AttemptNumber < 1 || req.AttemptNumber > 5 {
		return nil, errors.New("attempt_number must be between 1 and 5")
	}

	overallScore := computeOverallScore(
		req.ScoreIdeaAlignment,
		req.ScoreAngleMatch,
		req.ScoreHookMatch,
		req.ScorePersonaFit,
		req.ScoreVoiceDNA,
	)

	var adjustedPtr *string
	if req.AdjustedScript != "" {
		adjustedPtr = &req.AdjustedScript
	}

	attempt := &models.ScriptGenerationAttempt{
		ContentBriefID:     req.ContentBriefID,
		AttemptNumber:      req.AttemptNumber,
		ScriptBody:         req.ScriptBody,
		AdjustedScript:     adjustedPtr,
		ScoreIdeaAlignment: req.ScoreIdeaAlignment,
		ScoreAngleMatch:    req.ScoreAngleMatch,
		ScoreHookMatch:     req.ScoreHookMatch,
		ScorePersonaFit:    req.ScorePersonaFit,
		ScoreVoiceDNA:      req.ScoreVoiceDNA,
		OverallScore:       overallScore,
		IssuesFound:        req.IssuesFound,
		PassedThreshold:    overallScore >= models.ScriptPassThreshold,
	}

	if err := s.attemptRepo.Create(attempt); err != nil {
		return nil, err
	}

	// If this attempt passes, stop the loop immediately — do not waste further attempts
	if attempt.PassedThreshold {
		if err := s.attemptRepo.MarkSelected(attempt.ID); err != nil {
			return nil, err
		}
		if err := s.briefRepo.UpdateScriptSelection(
			req.ContentBriefID, nil, models.ScriptSourceWriterAgent,
			req.AttemptNumber, overallScore,
		); err != nil {
			return nil, err
		}
	}

	return attempt, nil
}

// SelectBestAttempt — called after 5 attempts without reaching 90%.
// Picks the highest-scoring attempt and marks it selected regardless of threshold.
func (s *pipelineService) SelectBestAttempt(briefID int32) (*models.ScriptGenerationAttempt, error) {
	attempts, err := s.attemptRepo.FindByBriefID(briefID)
	if err != nil || len(attempts) == 0 {
		return nil, errors.New("no attempts found for brief")
	}

	best := &attempts[0]
	for i := range attempts {
		if attempts[i].OverallScore > best.OverallScore {
			best = &attempts[i]
		}
	}

	if err := s.attemptRepo.MarkSelected(best.ID); err != nil {
		return nil, err
	}

	if err := s.briefRepo.UpdateScriptSelection(
		briefID, nil, models.ScriptSourceWriterAgent,
		int8(len(attempts)), best.OverallScore,
	); err != nil {
		return nil, err
	}

	return best, nil
}

// --- SSML stage ---

func (s *pipelineService) CreateSSMLScript(req CreateSSMLRequest) (*models.SSMLScript, error) {
	if req.SSMLBody == "" {
		return nil, errors.New("ssml_body is required")
	}

	var voiceIDPtr, platformPtr *string
	if req.ElevenLabsVoiceID != "" {
		voiceIDPtr = &req.ElevenLabsVoiceID
	}
	if req.Platform != "" {
		platformPtr = &req.Platform
	}

	ssml := &models.SSMLScript{
		ContentBriefID:    req.ContentBriefID,
		ApprovedScriptID:  req.ApprovedScriptID,
		SSMLBody:          req.SSMLBody,
		Platform:          platformPtr,
		ElevenLabsVoiceID: voiceIDPtr,
		AudioStatus:       models.AudioStatusPending,
	}
	if err := s.ssmlRepo.Create(ssml); err != nil {
		return nil, err
	}
	return ssml, nil
}

func (s *pipelineService) UpdateAudioStatus(ssmlID int32, status string, audioURL, jobID *string) error {
	return s.ssmlRepo.UpdateAudioStatus(ssmlID, status, audioURL, jobID)
}

func (s *pipelineService) MarkAudioReady(ssmlID int32, audioURL string, durationSeconds float64) error {
	if audioURL == "" {
		return errors.New("audio_url is required when marking audio ready")
	}
	return s.ssmlRepo.UpdateAudioReady(ssmlID, audioURL, durationSeconds)
}

// --- Agent confidence gating ---

// LogAgentConfidence records a confidence score and automatically flags it
// for human review if it falls below 70%, enforcing Nick's verification protocol.
func (s *pipelineService) LogAgentConfidence(req LogConfidenceRequest) (*models.AgentConfidenceLog, error) {
	if err := validateAgentName(req.AgentName); err != nil {
		return nil, err
	}

	flagged := req.ConfidenceScore < models.ConfidenceFlagThreshold

	log := &models.AgentConfidenceLog{
		AuditLogID:       req.AuditLogID,
		ContentBriefID:   req.ContentBriefID,
		AgentName:        req.AgentName,
		DecisionType:     req.DecisionType,
		ConfidenceScore:  req.ConfidenceScore,
		FlaggedForReview: flagged,
	}
	if err := s.confidenceRepo.Create(log); err != nil {
		return nil, err
	}
	return log, nil
}

func (s *pipelineService) ListFlaggedDecisions(resolved bool, page, pageSize int) ([]models.AgentConfidenceLog, int64, error) {
	return s.confidenceRepo.FindFlagged(resolved, page, pageSize)
}

func (s *pipelineService) ListEscalatedDecisions(page, pageSize int) ([]models.AgentConfidenceLog, int64, error) {
	return s.confidenceRepo.FindEscalated(page, pageSize)
}

func (s *pipelineService) ResolveConfidenceFlag(id int32) error {
	return s.confidenceRepo.MarkResolved(id)
}

// ---------------------------------------------------------------------------
// Private helpers
// ---------------------------------------------------------------------------

// computeOverallScore — equal-weighted average across five dimensions.
// VoiceDNA is a mandatory component: if it is zero, the script cannot pass 90%.
func computeOverallScore(ideaAlignment, angleMatch, hookMatch, personaFit, voiceDNA float64) float64 {
	return (ideaAlignment + angleMatch + hookMatch + personaFit + voiceDNA) / 5.0
}

// findBestBankedScript returns the highest validation_score and its script ID
// from the candidate list. Returns 0, 0 if the list is empty.
func findBestBankedScript(scripts []models.ApprovedScript) (float64, int32) {
	if len(scripts) == 0 {
		return 0, 0
	}
	best := scripts[0]
	for _, s := range scripts[1:] {
		if s.ValidationScore > best.ValidationScore {
			best = s
		}
	}
	return best.ValidationScore, best.ID
}

func validateBriefStatus(status string) error {
	valid := map[string]bool{
		models.BriefStatusPending:        true,
		models.BriefStatusAgent1Complete: true,
		models.BriefStatusAgent2Complete: true,
		models.BriefStatusScriptSelected: true,
		models.BriefStatusProduction:     true,
		models.BriefStatusReview:         true,
		models.BriefStatusPosted:         true,
		models.BriefStatusFailed:         true,
	}
	if !valid[status] {
		return fmt.Errorf("invalid brief status: %s", status)
	}
	return nil
}

func validateVideoType(videoType string) error {
	valid := map[string]bool{
		models.VideoTypeAvatar:   true,
		models.VideoTypeFaceless: true,
	}
	if !valid[videoType] {
		return errors.New("invalid video_type — must be avatar or faceless")
	}
	return nil
}

func validateAgentName(name string) error {
	valid := map[string]bool{
		models.AgentNameStrategic:          true,
		models.AgentNameCreative:           true,
		models.AgentNameValidation:         true,
		models.AgentNameWriter:             true,
		models.AgentNameSSML:               true,
		models.AgentNameAdversarialAuditor: true,
	}
	if !valid[name] {
		return fmt.Errorf("invalid agent_name: %s — must be from the closed agent list", name)
	}
	return nil
}
