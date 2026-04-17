package repositories

import (
	"backend-go/internal/models"

	"gorm.io/gorm"
)

// ===========================================================================
// ContentVideoRepository
// ===========================================================================

type ContentVideoRepository interface {
	Create(video *models.ContentVideo) error
	FindByID(id int32) (*models.ContentVideo, error)
	FindByContentBriefID(contentBriefID int32) (*models.ContentVideo, error)
	FindByRenderStatus(status string) ([]models.ContentVideo, error)
	UpdateRenderStatus(id int32, status string) error
	UpdateCoherenceCheck(id int32, passed bool, notes string) error
	UpdateVideoReady(id int32, videoURL string, durationSeconds float64, fileSizeBytes int64) error
	Update(video *models.ContentVideo) error
}

type contentVideoRepository struct {
	baseRepository
}

func NewContentVideoRepository(db *gorm.DB) ContentVideoRepository {
	return &contentVideoRepository{baseRepository{db: db}}
}

func (r *contentVideoRepository) Create(video *models.ContentVideo) error {
	return r.db.Create(video).Error
}

func (r *contentVideoRepository) FindByID(id int32) (*models.ContentVideo, error) {
	var video models.ContentVideo
	err := r.db.
		Preload("ContentBrief").
		Preload("SSMLScript").
		Preload("VideoScenes").
		First(&video, id).Error
	return &video, err
}

func (r *contentVideoRepository) FindByContentBriefID(contentBriefID int32) (*models.ContentVideo, error) {
	var video models.ContentVideo
	err := r.db.Where("content_brief_id = ?", contentBriefID).First(&video).Error
	return &video, err
}

func (r *contentVideoRepository) FindByRenderStatus(status string) ([]models.ContentVideo, error) {
	var videos []models.ContentVideo
	err := r.db.Where("render_status = ?", status).
		Order("created_at ASC").
		Find(&videos).Error
	return videos, err
}

func (r *contentVideoRepository) UpdateRenderStatus(id int32, status string) error {
	updates := map[string]interface{}{"render_status": status}
	if status == models.RenderStatusRendering {
		updates["render_started_at"] = gorm.Expr("NOW()")
	}
	return r.db.Model(&models.ContentVideo{}).Where("id = ?", id).Updates(updates).Error
}

func (r *contentVideoRepository) UpdateCoherenceCheck(id int32, passed bool, notes string) error {
	status := models.RenderStatusStitching
	if !passed {
		status = models.RenderStatusFailed
	}
	return r.db.Model(&models.ContentVideo{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"coherence_check_passed": passed,
			"coherence_check_notes":  notes,
			"render_status":          status,
		}).Error
}

func (r *contentVideoRepository) UpdateVideoReady(id int32, videoURL string, durationSeconds float64, fileSizeBytes int64) error {
	return r.db.Model(&models.ContentVideo{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"render_status":          models.RenderStatusReady,
			"video_url":              videoURL,
			"video_duration_seconds": durationSeconds,
			"file_size_bytes":        fileSizeBytes,
			"render_completed_at":    gorm.Expr("NOW()"),
		}).Error
}

func (r *contentVideoRepository) Update(video *models.ContentVideo) error {
	return r.db.Save(video).Error
}

// ===========================================================================
// VideoSceneRepository
// ===========================================================================

type VideoSceneRepository interface {
	CreateBatch(scenes []models.VideoScene) error
	FindByContentVideoID(contentVideoID int32) ([]models.VideoScene, error)
	FindByCoherenceStatus(contentVideoID int32, status string) ([]models.VideoScene, error)
	UpdateCoherenceStatus(id int32, status, failureReason string) error
	UpdateRenderStatus(id int32, status, sceneURL string) error
	IncrementRenderAttempts(id int32) error
	AllScenesCoherent(contentVideoID int32) (bool, error)
}

type videoSceneRepository struct {
	baseRepository
}

func NewVideoSceneRepository(db *gorm.DB) VideoSceneRepository {
	return &videoSceneRepository{baseRepository{db: db}}
}

func (r *videoSceneRepository) CreateBatch(scenes []models.VideoScene) error {
	return r.db.CreateInBatches(scenes, 10).Error
}

func (r *videoSceneRepository) FindByContentVideoID(contentVideoID int32) ([]models.VideoScene, error) {
	var scenes []models.VideoScene
	err := r.db.Where("content_video_id = ?", contentVideoID).
		Order("scene_number ASC").
		Find(&scenes).Error
	return scenes, err
}

func (r *videoSceneRepository) FindByCoherenceStatus(contentVideoID int32, status string) ([]models.VideoScene, error) {
	var scenes []models.VideoScene
	err := r.db.Where("content_video_id = ? AND coherence_status = ?", contentVideoID, status).
		Order("scene_number ASC").
		Find(&scenes).Error
	return scenes, err
}

func (r *videoSceneRepository) UpdateCoherenceStatus(id int32, status, failureReason string) error {
	return r.db.Model(&models.VideoScene{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"coherence_status":        status,
			"coherence_failure_reason": failureReason,
		}).Error
}

func (r *videoSceneRepository) UpdateRenderStatus(id int32, status, sceneURL string) error {
	updates := map[string]interface{}{"render_status": status}
	if sceneURL != "" {
		updates["scene_url"] = sceneURL
	}
	return r.db.Model(&models.VideoScene{}).Where("id = ?", id).Updates(updates).Error
}

func (r *videoSceneRepository) IncrementRenderAttempts(id int32) error {
	return r.db.Model(&models.VideoScene{}).
		Where("id = ?", id).
		Update("render_attempts", gorm.Expr("render_attempts + 1")).Error
}

// AllScenesCoherent returns true only when every scene for the video has passed coherence check.
func (r *videoSceneRepository) AllScenesCoherent(contentVideoID int32) (bool, error) {
	var failCount int64
	err := r.db.Model(&models.VideoScene{}).
		Where("content_video_id = ? AND coherence_status != ?", contentVideoID, models.CoherenceStatusPassed).
		Count(&failCount).Error
	return failCount == 0, err
}

// ===========================================================================
// PostingPackageRepository
// ===========================================================================

type PostingPackageRepository interface {
	Create(pkg *models.PostingPackage) error
	FindByID(id int32) (*models.PostingPackage, error)
	FindByTrackingID(trackingID string) (*models.PostingPackage, error)
	FindPendingReview(page, pageSize int) ([]models.PostingPackage, int64, error)
	FindByReviewStatus(status string, page, pageSize int) ([]models.PostingPackage, int64, error)
	UpdateReviewStatus(id int32, status string, reviewerUserID int32, rejectionReason *string) error
	MarkPosted(id int32, platformPostID string) error
	Update(pkg *models.PostingPackage) error
}

type postingPackageRepository struct {
	baseRepository
}

func NewPostingPackageRepository(db *gorm.DB) PostingPackageRepository {
	return &postingPackageRepository{baseRepository{db: db}}
}

func (r *postingPackageRepository) Create(pkg *models.PostingPackage) error {
	return r.db.Create(pkg).Error
}

func (r *postingPackageRepository) FindByID(id int32) (*models.PostingPackage, error) {
	var pkg models.PostingPackage
	err := r.db.
		Preload("ContentVideo").
		Preload("ContentBrief").
		Preload("Platform").
		Preload("ReviewedByUser").
		First(&pkg, id).Error
	return &pkg, err
}

func (r *postingPackageRepository) FindByTrackingID(trackingID string) (*models.PostingPackage, error) {
	var pkg models.PostingPackage
	err := r.db.Where("tracking_id = ?", trackingID).First(&pkg).Error
	return &pkg, err
}

func (r *postingPackageRepository) FindPendingReview(page, pageSize int) ([]models.PostingPackage, int64, error) {
	return r.FindByReviewStatus(models.ReviewStatusPendingReview, page, pageSize)
}

func (r *postingPackageRepository) FindByReviewStatus(status string, page, pageSize int) ([]models.PostingPackage, int64, error) {
	var pkgs []models.PostingPackage
	q := r.db.Model(&models.PostingPackage{}).Where("review_status = ?", status)
	total, err := countRows(q, &models.PostingPackage{})
	if err != nil {
		return nil, 0, err
	}
	err = r.paginate(page, pageSize).
		Where("review_status = ?", status).
		Preload("Platform").
		Order("created_at DESC").
		Find(&pkgs).Error
	return pkgs, total, err
}

func (r *postingPackageRepository) UpdateReviewStatus(id int32, status string, reviewerUserID int32, rejectionReason *string) error {
	updates := map[string]interface{}{
		"review_status":       status,
		"reviewed_by_user_id": reviewerUserID,
		"reviewed_at":         gorm.Expr("NOW()"),
	}
	if rejectionReason != nil {
		updates["rejection_reason"] = *rejectionReason
	}
	return r.db.Model(&models.PostingPackage{}).Where("id = ?", id).Updates(updates).Error
}

func (r *postingPackageRepository) MarkPosted(id int32, platformPostID string) error {
	return r.db.Model(&models.PostingPackage{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"review_status":    models.ReviewStatusPosted,
			"posted_at":        gorm.Expr("NOW()"),
			"platform_post_id": platformPostID,
		}).Error
}

func (r *postingPackageRepository) Update(pkg *models.PostingPackage) error {
	return r.db.Save(pkg).Error
}

// ===========================================================================
// VideoAnalyticRepository
// ===========================================================================

type VideoAnalyticRepository interface {
	Create(analytic *models.VideoAnalytic) error
	FindByPostingPackageID(postingPackageID int32) ([]models.VideoAnalytic, error)
	FindByWindow(postingPackageID int32, window string) (*models.VideoAnalytic, error)
	FindBeatingBaseline(window string, page, pageSize int) ([]models.VideoAnalytic, int64, error)
	CountBeatingBaseline(window string) (int64, error)
	FindLongformTriggers() ([]models.VideoAnalytic, error)
}

type videoAnalyticRepository struct {
	baseRepository
}

func NewVideoAnalyticRepository(db *gorm.DB) VideoAnalyticRepository {
	return &videoAnalyticRepository{baseRepository{db: db}}
}

func (r *videoAnalyticRepository) Create(analytic *models.VideoAnalytic) error {
	return r.db.Create(analytic).Error
}

func (r *videoAnalyticRepository) FindByPostingPackageID(postingPackageID int32) ([]models.VideoAnalytic, error) {
	var analytics []models.VideoAnalytic
	err := r.db.Where("posting_package_id = ?", postingPackageID).
		Order("collected_at DESC").
		Find(&analytics).Error
	return analytics, err
}

func (r *videoAnalyticRepository) FindByWindow(postingPackageID int32, window string) (*models.VideoAnalytic, error) {
	var analytic models.VideoAnalytic
	err := r.db.Where("posting_package_id = ? AND window = ?", postingPackageID, window).
		First(&analytic).Error
	return &analytic, err
}

func (r *videoAnalyticRepository) FindBeatingBaseline(window string, page, pageSize int) ([]models.VideoAnalytic, int64, error) {
	var analytics []models.VideoAnalytic
	q := r.db.Model(&models.VideoAnalytic{}).
		Where("window = ? AND beats_baseline = ?", window, true)
	total, err := countRows(q, &models.VideoAnalytic{})
	if err != nil {
		return nil, 0, err
	}
	err = r.paginate(page, pageSize).
		Where("window = ? AND beats_baseline = ?", window, true).
		Order("view_count DESC").
		Find(&analytics).Error
	return analytics, total, err
}

// CountBeatingBaseline — Phase 1 dashboard metric: how many videos beat the 38K baseline.
func (r *videoAnalyticRepository) CountBeatingBaseline(window string) (int64, error) {
	var count int64
	err := r.db.Model(&models.VideoAnalytic{}).
		Where("window = ? AND beats_baseline = ?", window, true).
		Count(&count).Error
	return count, err
}

// FindLongformTriggers — returns 7d analytics records flagged for long-form generation (Phase 2).
func (r *videoAnalyticRepository) FindLongformTriggers() ([]models.VideoAnalytic, error) {
	var analytics []models.VideoAnalytic
	err := r.db.Where("window = ? AND triggers_longform = ?", models.AnalyticsWindow7d, true).
		Order("view_count DESC").
		Find(&analytics).Error
	return analytics, err
}

// ===========================================================================
// AgentConfidenceLogRepository
// ===========================================================================

type AgentConfidenceLogRepository interface {
	Create(log *models.AgentConfidenceLog) error
	FindByContentBriefID(contentBriefID int32) ([]models.AgentConfidenceLog, error)
	FindFlagged(resolved bool, page, pageSize int) ([]models.AgentConfidenceLog, int64, error)
	FindEscalated(page, pageSize int) ([]models.AgentConfidenceLog, int64, error)
	MarkResolved(id int32) error
	UpdateEscalation(id int32, reason string) error
	IncrementRetry(id int32) error
	Update(log *models.AgentConfidenceLog) error
}

type agentConfidenceLogRepository struct {
	baseRepository
}

func NewAgentConfidenceLogRepository(db *gorm.DB) AgentConfidenceLogRepository {
	return &agentConfidenceLogRepository{baseRepository{db: db}}
}

func (r *agentConfidenceLogRepository) Create(log *models.AgentConfidenceLog) error {
	return r.db.Create(log).Error
}

func (r *agentConfidenceLogRepository) FindByContentBriefID(contentBriefID int32) ([]models.AgentConfidenceLog, error) {
	var logs []models.AgentConfidenceLog
	err := r.db.Where("content_brief_id = ?", contentBriefID).
		Order("created_at ASC").
		Find(&logs).Error
	return logs, err
}

func (r *agentConfidenceLogRepository) FindFlagged(resolved bool, page, pageSize int) ([]models.AgentConfidenceLog, int64, error) {
	var logs []models.AgentConfidenceLog
	q := r.db.Model(&models.AgentConfidenceLog{}).
		Where("flagged_for_review = ? AND resolved = ?", true, resolved)
	total, err := countRows(q, &models.AgentConfidenceLog{})
	if err != nil {
		return nil, 0, err
	}
	err = r.paginate(page, pageSize).
		Where("flagged_for_review = ? AND resolved = ?", true, resolved).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, total, err
}

func (r *agentConfidenceLogRepository) FindEscalated(page, pageSize int) ([]models.AgentConfidenceLog, int64, error) {
	var logs []models.AgentConfidenceLog
	q := r.db.Model(&models.AgentConfidenceLog{}).
		Where("escalated = ? AND resolved = ?", true, false)
	total, err := countRows(q, &models.AgentConfidenceLog{})
	if err != nil {
		return nil, 0, err
	}
	err = r.paginate(page, pageSize).
		Where("escalated = ? AND resolved = ?", true, false).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, total, err
}

func (r *agentConfidenceLogRepository) MarkResolved(id int32) error {
	return r.db.Model(&models.AgentConfidenceLog{}).
		Where("id = ?", id).
		Update("resolved", true).Error
}

func (r *agentConfidenceLogRepository) UpdateEscalation(id int32, reason string) error {
	return r.db.Model(&models.AgentConfidenceLog{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"escalated":         true,
			"escalation_reason": reason,
		}).Error
}

func (r *agentConfidenceLogRepository) IncrementRetry(id int32) error {
	return r.db.Model(&models.AgentConfidenceLog{}).
		Where("id = ?", id).
		Update("retry_count", gorm.Expr("retry_count + 1")).Error
}

func (r *agentConfidenceLogRepository) Update(log *models.AgentConfidenceLog) error {
	return r.db.Save(log).Error
}
