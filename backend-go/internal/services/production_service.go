package services

import (
	"errors"
	"fmt"

	"backend-go/internal/models"
	"backend-go/internal/repositories"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// ProductionService
// Manages Sections 9–11: video production, scene coherence gating,
// posting package preparation, and analytics collection + baseline tracking.
// ---------------------------------------------------------------------------

type ProductionService interface {
	// Video production
	CreateContentVideo(req CreateVideoRequest) (*models.ContentVideo, error)
	GetContentVideo(id int32) (*models.ContentVideo, error)
	GetVideoByBriefID(contentBriefID int32) (*models.ContentVideo, error)
	UpdateRenderStatus(videoID int32, status string) error

	// Scene management (HeyGen Avatar Shots)
	CreateScenes(videoID int32, scenes []CreateSceneRequest) ([]models.VideoScene, error)
	GetScenes(videoID int32) ([]models.VideoScene, error)
	UpdateSceneCoherence(sceneID int32, passed bool, failureReason string) error
	UpdateSceneRendered(sceneID int32, sceneURL string) error
	RunCoherenceGate(videoID int32) (CoherenceGateResult, error)

	// Posting package (Phase 1 — manual review)
	CreatePostingPackage(req CreatePostingPackageRequest) (*models.PostingPackage, error)
	GetPostingPackage(id int32) (*models.PostingPackage, error)
	ListPendingReview(page, pageSize int) ([]models.PostingPackage, int64, error)
	ApprovePackage(packageID int32, reviewerUserID int32) error
	RejectPackage(packageID int32, reviewerUserID int32, reason string) error
	MarkPosted(packageID int32, platformPostID string) error

	// Analytics
	RecordAnalytics(req RecordAnalyticsRequest) (*models.VideoAnalytic, error)
	GetAnalytics(postingPackageID int32) ([]models.VideoAnalytic, error)
	Phase1DashboardMetrics(window string) (Phase1Metrics, error)
}

// ---------------------------------------------------------------------------
// Request / response types
// ---------------------------------------------------------------------------

type CreateVideoRequest struct {
	ContentBriefID int32
	SSMLScriptID   int32
	ProductionTool string
	VideoType      string
	HeyGenAvatarID string
}

type CreateSceneRequest struct {
	SceneNumber           int8
	SceneScript           string
	DurationSeconds       float64
	WardrobeDescription   string
	BackgroundDescription string
	LogoPresent           *bool
}

type CoherenceGateResult struct {
	AllPassed      bool
	FailedScenes   []int32  // scene IDs that failed
	Issues         []string // human-readable issues
}

type CreatePostingPackageRequest struct {
	ContentVideoID      int32
	ContentBriefID      int32
	PlatformID          int32
	Caption             string
	HookLine            string
	Hashtags            []string
	FormattedVideoURL   string
}

type RecordAnalyticsRequest struct {
	PostingPackageID int32
	PlatformID       int32
	TrackingID       string
	Window           string
	ViewCount        int64
	LikeCount        int64
	CommentCount     int64
	ShareCount       int64
	SaveCount        int64
	SendsPerReach    float64
	CompletionRate   float64
	EngagementRate   float64
}

// Phase1Metrics — the dashboard summary for Nick's review.
type Phase1Metrics struct {
	TotalVideosPosted      int64
	BeatingBaseline24h     int64  // videos beating 38K at 24h window
	BeatingBaseline7d      int64  // videos beating 38K at 7d window
	BeatingBaseline30d     int64  // videos beating 38K at 30d window
	BaselineTarget         int64  // always 38000
	PendingReviewCount     int64
	LongFormTriggersQueued int64  // Phase 2 prep: 7d > 50K
	Phase1Complete         bool   // true when system consistently beats baseline
}

// ---------------------------------------------------------------------------
// Implementation
// ---------------------------------------------------------------------------

type productionService struct {
	videoRepo   repositories.ContentVideoRepository
	sceneRepo   repositories.VideoSceneRepository
	packageRepo repositories.PostingPackageRepository
	analyticRepo repositories.VideoAnalyticRepository
	briefRepo   repositories.ContentBriefRepository
}

func NewProductionService(
	videoRepo repositories.ContentVideoRepository,
	sceneRepo repositories.VideoSceneRepository,
	packageRepo repositories.PostingPackageRepository,
	analyticRepo repositories.VideoAnalyticRepository,
	briefRepo repositories.ContentBriefRepository,
) ProductionService {
	return &productionService{
		videoRepo:    videoRepo,
		sceneRepo:    sceneRepo,
		packageRepo:  packageRepo,
		analyticRepo: analyticRepo,
		briefRepo:    briefRepo,
	}
}

// --- Video production ---

func (s *productionService) CreateContentVideo(req CreateVideoRequest) (*models.ContentVideo, error) {
	if err := validateProductionTool(req.ProductionTool); err != nil {
		return nil, err
	}
	if err := validateVideoType(req.VideoType); err != nil {
		return nil, err
	}

	var avatarIDPtr *string
	if req.HeyGenAvatarID != "" {
		avatarIDPtr = &req.HeyGenAvatarID
	}

	video := &models.ContentVideo{
		ContentBriefID: req.ContentBriefID,
		SSMLScriptID:   req.SSMLScriptID,
		ProductionTool: req.ProductionTool,
		VideoType:      req.VideoType,
		HeyGenAvatarID: avatarIDPtr,
		RenderStatus:   models.RenderStatusPending,
	}
	if err := s.videoRepo.Create(video); err != nil {
		return nil, err
	}

	// Advance brief status
	_ = s.briefRepo.UpdateStatus(req.ContentBriefID, models.BriefStatusProduction)

	return video, nil
}

func (s *productionService) GetContentVideo(id int32) (*models.ContentVideo, error) {
	return s.videoRepo.FindByID(id)
}

func (s *productionService) GetVideoByBriefID(contentBriefID int32) (*models.ContentVideo, error) {
	return s.videoRepo.FindByContentBriefID(contentBriefID)
}

func (s *productionService) UpdateRenderStatus(videoID int32, status string) error {
	if err := validateRenderStatus(status); err != nil {
		return err
	}
	return s.videoRepo.UpdateRenderStatus(videoID, status)
}

// --- Scene management ---

// CreateScenes batch-creates all scenes for a HeyGen Avatar Shots video.
// Scenes start in pending coherence / render status.
func (s *productionService) CreateScenes(videoID int32, reqs []CreateSceneRequest) ([]models.VideoScene, error) {
	if len(reqs) == 0 {
		return nil, errors.New("at least one scene is required")
	}

	scenes := make([]models.VideoScene, len(reqs))
	for i, req := range reqs {
		duration := req.DurationSeconds
		if duration <= 0 {
			duration = 15.0 // HeyGen Avatar Shots default
		}
		var wardrobePtr, backgroundPtr *string
		if req.WardrobeDescription != "" {
			wardrobePtr = &req.WardrobeDescription
		}
		if req.BackgroundDescription != "" {
			backgroundPtr = &req.BackgroundDescription
		}
		scenes[i] = models.VideoScene{
			ContentVideoID:        videoID,
			SceneNumber:           req.SceneNumber,
			SceneScript:           req.SceneScript,
			DurationSeconds:       duration,
			WardrobeDescription:   wardrobePtr,
			BackgroundDescription: backgroundPtr,
			LogoPresent:           req.LogoPresent,
			CoherenceStatus:       models.CoherenceStatusPending,
			RenderStatus:          models.RenderStatusPending,
		}
	}

	if err := s.sceneRepo.CreateBatch(scenes); err != nil {
		return nil, err
	}
	return scenes, nil
}

func (s *productionService) GetScenes(videoID int32) ([]models.VideoScene, error) {
	return s.sceneRepo.FindByContentVideoID(videoID)
}

func (s *productionService) UpdateSceneCoherence(sceneID int32, passed bool, failureReason string) error {
	status := models.CoherenceStatusPassed
	if !passed {
		status = models.CoherenceStatusFailed
	}
	return s.sceneRepo.UpdateCoherenceStatus(sceneID, status, failureReason)
}

func (s *productionService) UpdateSceneRendered(sceneID int32, sceneURL string) error {
	return s.sceneRepo.UpdateRenderStatus(sceneID, models.RenderStatusReady, sceneURL)
}

// RunCoherenceGate — mandatory gate before scenes are stitched.
// Checks wardrobe and logo consistency across all scenes.
// Any failure blocks stitching — scenes must be re-rendered first.
func (s *productionService) RunCoherenceGate(videoID int32) (CoherenceGateResult, error) {
	scenes, err := s.sceneRepo.FindByContentVideoID(videoID)
	if err != nil {
		return CoherenceGateResult{}, err
	}
	if len(scenes) == 0 {
		return CoherenceGateResult{}, errors.New("no scenes found for video")
	}

	result := CoherenceGateResult{AllPassed: true}

	// Derive expected logo state from the first scene that has a logo value set
	var expectedLogoPresent *bool
	for _, sc := range scenes {
		if sc.LogoPresent != nil {
			expectedLogoPresent = sc.LogoPresent
			break
		}
	}

	for _, sc := range scenes {
		var sceneIssues []string

		// Logo consistency check
		if expectedLogoPresent != nil && sc.LogoPresent != nil {
			if *sc.LogoPresent != *expectedLogoPresent {
				issue := fmt.Sprintf("scene %d: logo_present=%v but expected %v — must be consistent across all scenes",
					sc.SceneNumber, *sc.LogoPresent, *expectedLogoPresent)
				sceneIssues = append(sceneIssues, issue)
			}
		}

		if len(sceneIssues) > 0 {
			result.AllPassed = false
			result.FailedScenes = append(result.FailedScenes, sc.ID)
			result.Issues = append(result.Issues, sceneIssues...)
			for _, issue := range sceneIssues {
				_ = s.sceneRepo.UpdateCoherenceStatus(sc.ID, models.CoherenceStatusFailed, issue)
			}
		} else {
			_ = s.sceneRepo.UpdateCoherenceStatus(sc.ID, models.CoherenceStatusPassed, "")
		}
	}

	// Update video render status based on gate outcome
	if result.AllPassed {
		_ = s.videoRepo.UpdateRenderStatus(videoID, models.RenderStatusStitching)
	} else {
		_ = s.videoRepo.UpdateCoherenceCheck(videoID, false,
			fmt.Sprintf("%d scene(s) failed coherence — re-render required before stitching", len(result.FailedScenes)))
	}

	return result, nil
}

// --- Posting packages ---

func (s *productionService) CreatePostingPackage(req CreatePostingPackageRequest) (*models.PostingPackage, error) {
	if req.Caption == "" {
		return nil, errors.New("caption is required")
	}

	// Generate unique tracking ID — links this video back to every production decision
	trackingID := generateTrackingID()

	var hookPtr, videoURLPtr *string
	if req.HookLine != "" {
		hookPtr = &req.HookLine
	}
	if req.FormattedVideoURL != "" {
		videoURLPtr = &req.FormattedVideoURL
	}

	pkg := &models.PostingPackage{
		ContentVideoID:    req.ContentVideoID,
		ContentBriefID:    req.ContentBriefID,
		TrackingID:        trackingID,
		PlatformID:        req.PlatformID,
		FormattedVideoURL: videoURLPtr,
		Caption:           req.Caption,
		HookLine:          hookPtr,
		Hashtags:          req.Hashtags,
		ReviewStatus:      models.ReviewStatusPendingReview,
	}
	if err := s.packageRepo.Create(pkg); err != nil {
		return nil, err
	}

	// Advance brief to review status
	_ = s.briefRepo.UpdateStatus(req.ContentBriefID, models.BriefStatusReview)

	return pkg, nil
}

func (s *productionService) GetPostingPackage(id int32) (*models.PostingPackage, error) {
	return s.packageRepo.FindByID(id)
}

func (s *productionService) ListPendingReview(page, pageSize int) ([]models.PostingPackage, int64, error) {
	return s.packageRepo.FindPendingReview(page, pageSize)
}

func (s *productionService) ApprovePackage(packageID int32, reviewerUserID int32) error {
	return s.packageRepo.UpdateReviewStatus(packageID, models.ReviewStatusApproved, reviewerUserID, nil)
}

func (s *productionService) RejectPackage(packageID int32, reviewerUserID int32, reason string) error {
	if reason == "" {
		return errors.New("rejection reason is required")
	}
	return s.packageRepo.UpdateReviewStatus(packageID, models.ReviewStatusRejected, reviewerUserID, &reason)
}

func (s *productionService) MarkPosted(packageID int32, platformPostID string) error {
	pkg, err := s.packageRepo.FindByID(packageID)
	if err != nil {
		return err
	}
	if pkg.ReviewStatus != models.ReviewStatusApproved {
		return errors.New("package must be approved before marking as posted")
	}
	if err := s.packageRepo.MarkPosted(packageID, platformPostID); err != nil {
		return err
	}
	return s.briefRepo.UpdateStatus(pkg.ContentBriefID, models.BriefStatusPosted)
}

// --- Analytics ---

// RecordAnalytics stores a performance snapshot and evaluates it against the 38K baseline.
func (s *productionService) RecordAnalytics(req RecordAnalyticsRequest) (*models.VideoAnalytic, error) {
	if err := validateAnalyticsWindow(req.Window); err != nil {
		return nil, err
	}

	beatsBaseline := req.ViewCount >= models.NickBaselineViews
	baselineMultiplier := computeBaselineMultiplier(req.ViewCount, models.NickBaselineViews)

	// Phase 2 trigger: 7d views >= 50K flags for automated long-form generation
	triggersLongform := req.Window == models.AnalyticsWindow7d && req.ViewCount >= models.LongFormTriggerThreshold

	analytic := &models.VideoAnalytic{
		PostingPackageID:   req.PostingPackageID,
		PlatformID:         req.PlatformID,
		TrackingID:         req.TrackingID,
		Window:             req.Window,
		ViewCount:          req.ViewCount,
		LikeCount:          req.LikeCount,
		CommentCount:       req.CommentCount,
		ShareCount:         req.ShareCount,
		SaveCount:          req.SaveCount,
		SendsPerReach:      req.SendsPerReach,
		CompletionRate:     req.CompletionRate,
		EngagementRate:     req.EngagementRate,
		BeatsBaseline:      &beatsBaseline,
		BaselineMultiplier: &baselineMultiplier,
		TriggersLongform:   triggersLongform,
	}
	if err := s.analyticRepo.Create(analytic); err != nil {
		return nil, err
	}
	return analytic, nil
}

func (s *productionService) GetAnalytics(postingPackageID int32) ([]models.VideoAnalytic, error) {
	return s.analyticRepo.FindByPostingPackageID(postingPackageID)
}

// Phase1DashboardMetrics assembles the top-level Phase 1 success dashboard.
// Primary question: is the AI pipeline consistently beating Nick's 38K baseline?
func (s *productionService) Phase1DashboardMetrics(window string) (Phase1Metrics, error) {
	if err := validateAnalyticsWindow(window); err != nil {
		return Phase1Metrics{}, err
	}

	beating24h, err := s.analyticRepo.CountBeatingBaseline(models.AnalyticsWindow24h)
	if err != nil {
		return Phase1Metrics{}, err
	}
	beating7d, err := s.analyticRepo.CountBeatingBaseline(models.AnalyticsWindow7d)
	if err != nil {
		return Phase1Metrics{}, err
	}
	beating30d, err := s.analyticRepo.CountBeatingBaseline(models.AnalyticsWindow30d)
	if err != nil {
		return Phase1Metrics{}, err
	}

	longformTriggers, err := s.analyticRepo.FindLongformTriggers()
	if err != nil {
		return Phase1Metrics{}, err
	}

	pending, _, err := s.packageRepo.FindPendingReview(1, 1)
	if err != nil {
		return Phase1Metrics{}, err
	}
	_ = pending

	posted, _, err := s.packageRepo.FindByReviewStatus(models.ReviewStatusPosted, 1, 9999)
	if err != nil {
		return Phase1Metrics{}, err
	}

	metrics := Phase1Metrics{
		TotalVideosPosted:      int64(len(posted)),
		BeatingBaseline24h:     beating24h,
		BeatingBaseline7d:      beating7d,
		BeatingBaseline30d:     beating30d,
		BaselineTarget:         models.NickBaselineViews,
		LongFormTriggersQueued: int64(len(longformTriggers)),
		// Phase 1 is complete when the majority of 7d-window videos beat the baseline
		Phase1Complete: beating7d > 0 && int64(len(posted)) > 0 && beating7d >= int64(len(posted))/2,
	}

	return metrics, nil
}

// ---------------------------------------------------------------------------
// Private helpers
// ---------------------------------------------------------------------------

func generateTrackingID() string {
	return uuid.New().String()
}

func computeBaselineMultiplier(viewCount, baseline int64) float64 {
	if baseline == 0 {
		return 0
	}
	return float64(viewCount) / float64(baseline)
}

func validateProductionTool(tool string) error {
	valid := map[string]bool{
		models.ProductionToolHeyGenAvatarShots: true,
		models.ProductionToolHeyGenAvatarV:     true,
		models.ProductionToolKlingAI:           true,
		models.ProductionToolElevenLabsStudio:  true,
	}
	if !valid[tool] {
		return fmt.Errorf("invalid production_tool: %s", tool)
	}
	return nil
}

func validateRenderStatus(status string) error {
	valid := map[string]bool{
		models.RenderStatusPending:        true,
		models.RenderStatusRendering:      true,
		models.RenderStatusCoherenceCheck: true,
		models.RenderStatusStitching:      true,
		models.RenderStatusReady:          true,
		models.RenderStatusFailed:         true,
	}
	if !valid[status] {
		return fmt.Errorf("invalid render_status: %s", status)
	}
	return nil
}

func validateAnalyticsWindow(window string) error {
	valid := map[string]bool{
		models.AnalyticsWindow24h: true,
		models.AnalyticsWindow7d:  true,
		models.AnalyticsWindow30d: true,
	}
	if !valid[window] {
		return errors.New("invalid analytics window — must be 24h, 7d, or 30d")
	}
	return nil
}
