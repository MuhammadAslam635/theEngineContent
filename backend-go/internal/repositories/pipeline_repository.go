package repositories

import (
	"backend-go/internal/models"

	"gorm.io/gorm"
)

// ===========================================================================
// ContentBriefRepository
// ===========================================================================

type ContentBriefRepository interface {
	Create(brief *models.ContentBrief) error
	FindByID(id int32) (*models.ContentBrief, error)
	FindByStatus(status string, page, pageSize int) ([]models.ContentBrief, int64, error)
	FindByContentType(contentType string, page, pageSize int) ([]models.ContentBrief, int64, error)
	UpdateStatus(id int32, status string) error
	UpdateAgent1Output(id int32, platformID int32, contentType, videoType string, personaID int32, trendSignalID, outlierReelRefID *int32, confidence float64, reasoning string) error
	UpdateAgent2Output(id int32, primaryAngleID, secondaryAngleID, primaryHookID *int32, confidence float64, reasoning string, avoidWhenVerified bool) error
	UpdateScriptSelection(id int32, scriptID *int32, method string, attempts int8, score float64) error
	Update(brief *models.ContentBrief) error
}

type contentBriefRepository struct {
	baseRepository
}

func NewContentBriefRepository(db *gorm.DB) ContentBriefRepository {
	return &contentBriefRepository{baseRepository{db: db}}
}

func (r *contentBriefRepository) Create(brief *models.ContentBrief) error {
	return r.db.Create(brief).Error
}

func (r *contentBriefRepository) FindByID(id int32) (*models.ContentBrief, error) {
	var brief models.ContentBrief
	err := r.db.
		Preload("PrimaryPlatform").
		Preload("Persona").
		Preload("TrendSignal").
		Preload("OutlierReelRef").
		Preload("PrimaryAngle").
		Preload("SecondaryAngle").
		Preload("PrimaryHook").
		Preload("SelectedScript").
		First(&brief, id).Error
	return &brief, err
}

func (r *contentBriefRepository) FindByStatus(status string, page, pageSize int) ([]models.ContentBrief, int64, error) {
	var briefs []models.ContentBrief
	q := applyStatusFilter(r.db, status)
	total, err := countRows(q.Model(&models.ContentBrief{}), &models.ContentBrief{})
	if err != nil {
		return nil, 0, err
	}
	err = r.paginate(page, pageSize).
		Where("status = ?", status).
		Order("created_at DESC").
		Find(&briefs).Error
	return briefs, total, err
}

func (r *contentBriefRepository) FindByContentType(contentType string, page, pageSize int) ([]models.ContentBrief, int64, error) {
	var briefs []models.ContentBrief
	q := r.db.Model(&models.ContentBrief{}).Where("content_type = ?", contentType)
	total, err := countRows(q, &models.ContentBrief{})
	if err != nil {
		return nil, 0, err
	}
	err = r.paginate(page, pageSize).
		Where("content_type = ?", contentType).
		Order("created_at DESC").
		Find(&briefs).Error
	return briefs, total, err
}

func (r *contentBriefRepository) UpdateStatus(id int32, status string) error {
	return r.db.Model(&models.ContentBrief{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *contentBriefRepository) UpdateAgent1Output(
	id int32,
	platformID int32,
	contentType, videoType string,
	personaID int32,
	trendSignalID, outlierReelRefID *int32,
	confidence float64,
	reasoning string,
) error {
	return r.db.Model(&models.ContentBrief{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"primary_platform_id":  platformID,
			"content_type":         contentType,
			"video_type":           videoType,
			"persona_id":           personaID,
			"trend_signal_id":      trendSignalID,
			"outlier_reel_ref_id":  outlierReelRefID,
			"agent1_confidence":    confidence,
			"agent1_reasoning":     reasoning,
			"status":               models.BriefStatusAgent1Complete,
		}).Error
}

func (r *contentBriefRepository) UpdateAgent2Output(
	id int32,
	primaryAngleID, secondaryAngleID, primaryHookID *int32,
	confidence float64,
	reasoning string,
	avoidWhenVerified bool,
) error {
	return r.db.Model(&models.ContentBrief{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"primary_angle_id":     primaryAngleID,
			"secondary_angle_id":   secondaryAngleID,
			"primary_hook_id":      primaryHookID,
			"agent2_confidence":    confidence,
			"agent2_reasoning":     reasoning,
			"avoid_when_verified":  avoidWhenVerified,
			"status":               models.BriefStatusAgent2Complete,
		}).Error
}

func (r *contentBriefRepository) UpdateScriptSelection(
	id int32,
	scriptID *int32,
	method string,
	attempts int8,
	score float64,
) error {
	return r.db.Model(&models.ContentBrief{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"selected_script_id":      scriptID,
			"script_selection_method": method,
			"script_attempts":         attempts,
			"final_script_score":      score,
			"status":                  models.BriefStatusScriptSelected,
		}).Error
}

func (r *contentBriefRepository) Update(brief *models.ContentBrief) error {
	return r.db.Save(brief).Error
}

// ===========================================================================
// ScriptGenerationAttemptRepository
// ===========================================================================

type ScriptGenerationAttemptRepository interface {
	Create(attempt *models.ScriptGenerationAttempt) error
	FindByBriefID(contentBriefID int32) ([]models.ScriptGenerationAttempt, error)
	FindSelectedByBriefID(contentBriefID int32) (*models.ScriptGenerationAttempt, error)
	CountByBriefID(contentBriefID int32) (int64, error)
	MarkSelected(id int32) error
	Update(attempt *models.ScriptGenerationAttempt) error
}

type scriptGenerationAttemptRepository struct {
	baseRepository
}

func NewScriptGenerationAttemptRepository(db *gorm.DB) ScriptGenerationAttemptRepository {
	return &scriptGenerationAttemptRepository{baseRepository{db: db}}
}

func (r *scriptGenerationAttemptRepository) Create(attempt *models.ScriptGenerationAttempt) error {
	return r.db.Create(attempt).Error
}

func (r *scriptGenerationAttemptRepository) FindByBriefID(contentBriefID int32) ([]models.ScriptGenerationAttempt, error) {
	var attempts []models.ScriptGenerationAttempt
	err := r.db.Where("content_brief_id = ?", contentBriefID).
		Order("attempt_number ASC").
		Find(&attempts).Error
	return attempts, err
}

func (r *scriptGenerationAttemptRepository) FindSelectedByBriefID(contentBriefID int32) (*models.ScriptGenerationAttempt, error) {
	var attempt models.ScriptGenerationAttempt
	err := r.db.Where("content_brief_id = ? AND selected = ?", contentBriefID, true).
		First(&attempt).Error
	return &attempt, err
}

func (r *scriptGenerationAttemptRepository) CountByBriefID(contentBriefID int32) (int64, error) {
	var count int64
	err := r.db.Model(&models.ScriptGenerationAttempt{}).
		Where("content_brief_id = ?", contentBriefID).
		Count(&count).Error
	return count, err
}

// MarkSelected clears any previous selection for this brief then marks the given attempt.
func (r *scriptGenerationAttemptRepository) MarkSelected(id int32) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Get the brief_id for this attempt
		var attempt models.ScriptGenerationAttempt
		if err := tx.First(&attempt, id).Error; err != nil {
			return err
		}
		// Clear all previous selections for the same brief
		if err := tx.Model(&models.ScriptGenerationAttempt{}).
			Where("content_brief_id = ? AND selected = ?", attempt.ContentBriefID, true).
			Update("selected", false).Error; err != nil {
			return err
		}
		// Mark the target attempt
		return tx.Model(&models.ScriptGenerationAttempt{}).
			Where("id = ?", id).
			Update("selected", true).Error
	})
}

func (r *scriptGenerationAttemptRepository) Update(attempt *models.ScriptGenerationAttempt) error {
	return r.db.Save(attempt).Error
}

// ===========================================================================
// SSMLScriptRepository
// ===========================================================================

type SSMLScriptRepository interface {
	Create(ssml *models.SSMLScript) error
	FindByID(id int32) (*models.SSMLScript, error)
	FindByContentBriefID(contentBriefID int32) (*models.SSMLScript, error)
	UpdateAudioStatus(id int32, status string, audioURL, jobID *string) error
	UpdateAudioReady(id int32, audioURL string, durationSeconds float64) error
	Update(ssml *models.SSMLScript) error
}

type ssmlScriptRepository struct {
	baseRepository
}

func NewSSMLScriptRepository(db *gorm.DB) SSMLScriptRepository {
	return &ssmlScriptRepository{baseRepository{db: db}}
}

func (r *ssmlScriptRepository) Create(ssml *models.SSMLScript) error {
	return r.db.Create(ssml).Error
}

func (r *ssmlScriptRepository) FindByID(id int32) (*models.SSMLScript, error) {
	var ssml models.SSMLScript
	err := r.db.Preload("ContentBrief").Preload("ApprovedScript").First(&ssml, id).Error
	return &ssml, err
}

func (r *ssmlScriptRepository) FindByContentBriefID(contentBriefID int32) (*models.SSMLScript, error) {
	var ssml models.SSMLScript
	err := r.db.Where("content_brief_id = ?", contentBriefID).First(&ssml).Error
	return &ssml, err
}

func (r *ssmlScriptRepository) UpdateAudioStatus(id int32, status string, audioURL, jobID *string) error {
	return r.db.Model(&models.SSMLScript{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"audio_status":         status,
			"elevenlabs_audio_url": audioURL,
			"elevenlabs_job_id":    jobID,
			"submitted_at":         gorm.Expr("NOW()"),
		}).Error
}

func (r *ssmlScriptRepository) UpdateAudioReady(id int32, audioURL string, durationSeconds float64) error {
	return r.db.Model(&models.SSMLScript{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"audio_status":           models.AudioStatusReady,
			"elevenlabs_audio_url":   audioURL,
			"audio_duration_seconds": durationSeconds,
			"audio_ready_at":         gorm.Expr("NOW()"),
		}).Error
}

func (r *ssmlScriptRepository) Update(ssml *models.SSMLScript) error {
	return r.db.Save(ssml).Error
}
