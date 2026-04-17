package services

import (
	"errors"
	"time"

	"backend-go/internal/models"
	"backend-go/internal/repositories"
)

// ---------------------------------------------------------------------------
// IntelligenceService
// Handles all Section 3 logic: outlier reel detection and trend signal management.
// The 5X threshold is account-relative — computed per competitor account.
// ---------------------------------------------------------------------------

type IntelligenceService interface {
	// Competitor accounts
	AddCompetitorAccount(platformID int32, handle, displayName, profileURL, niche string) (*models.CompetitorAccount, error)
	ListActiveCompetitorAccounts() ([]models.CompetitorAccount, error)
	ListByPlatform(platformID int32) ([]models.CompetitorAccount, error)
	DeactivateCompetitorAccount(id int32) error

	// Outlier reel ingestion (called by Scout / SociaVault webhook)
	IngestReel(req IngestReelRequest) (*models.OutlierReel, error)
	UpdateAccountBaseline(accountID int32, newAvgViewCount int64) error

	// Relevance filter
	ListPendingRelevanceFilter(limit int) ([]models.OutlierReel, error)
	ApproveReel(id int32, reason string) error
	RejectReel(id int32, reason string) error

	// Trend signals
	UpsertTrendSignal(req UpsertTrendSignalRequest) (*models.TrendSignal, error)
	ListActionableTrendSignals() ([]models.TrendSignal, error)
	GetTrendSignal(id int32) (*models.TrendSignal, error)
}

// ---------------------------------------------------------------------------
// Request types
// ---------------------------------------------------------------------------

type IngestReelRequest struct {
	CompetitorAccountID int32
	PlatformID          int32
	ExternalReelID      string
	Title               string
	URL                 string
	ViewCount           int64
	LikeCount           int64
	CommentCount        int64
	ShareCount          int64
	RawTranscript       string
	PublishedAt         *time.Time
}

type UpsertTrendSignalRequest struct {
	Topic         string
	PlatformID    int32
	Volume30d     int
	Volume7d      int
	Volume24h     int
	Stage         string
	UrgencyRating int8
}

// ---------------------------------------------------------------------------
// Implementation
// ---------------------------------------------------------------------------

type intelligenceService struct {
	competitorRepo  repositories.CompetitorAccountRepository
	outlierReelRepo repositories.OutlierReelRepository
	trendSignalRepo repositories.TrendSignalRepository
	platformRepo    repositories.PlatformRepository
}

func NewIntelligenceService(
	competitorRepo repositories.CompetitorAccountRepository,
	outlierReelRepo repositories.OutlierReelRepository,
	trendSignalRepo repositories.TrendSignalRepository,
	platformRepo repositories.PlatformRepository,
) IntelligenceService {
	return &intelligenceService{
		competitorRepo:  competitorRepo,
		outlierReelRepo: outlierReelRepo,
		trendSignalRepo: trendSignalRepo,
		platformRepo:    platformRepo,
	}
}

// --- Competitor accounts ---

func (s *intelligenceService) AddCompetitorAccount(platformID int32, handle, displayName, profileURL, niche string) (*models.CompetitorAccount, error) {
	if _, err := s.platformRepo.FindByID(platformID); err != nil {
		return nil, errors.New("platform not found")
	}
	var profileURLPtr *string
	if profileURL != "" {
		profileURLPtr = &profileURL
	}
	account := &models.CompetitorAccount{
		PlatformID:  platformID,
		Handle:      handle,
		DisplayName: displayName,
		ProfileURL:  profileURLPtr,
		Niche:       niche,
	}
	if err := s.competitorRepo.Create(account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *intelligenceService) ListActiveCompetitorAccounts() ([]models.CompetitorAccount, error) {
	return s.competitorRepo.FindAllActive()
}

func (s *intelligenceService) ListByPlatform(platformID int32) ([]models.CompetitorAccount, error) {
	return s.competitorRepo.FindByPlatform(platformID)
}

func (s *intelligenceService) DeactivateCompetitorAccount(id int32) error {
	return s.competitorRepo.Delete(id)
}

// --- Outlier reel ingestion ---

// IngestReel receives a reel from Scout, checks it against the 5X threshold,
// and stores it if it qualifies. The threshold is account-relative.
func (s *intelligenceService) IngestReel(req IngestReelRequest) (*models.OutlierReel, error) {
	account, err := s.competitorRepo.FindByID(req.CompetitorAccountID)
	if err != nil {
		return nil, errors.New("competitor account not found")
	}

	// Idempotency — skip if reel already ingested
	existing, _ := s.outlierReelRepo.FindByExternalID(req.PlatformID, req.ExternalReelID)
	if existing != nil {
		return existing, nil
	}

	// 5X threshold check — account-relative, not a global fixed number
	if account.OutlierThreshold > 0 && req.ViewCount < account.OutlierThreshold {
		return nil, errors.New("reel does not meet 5X outlier threshold for this account")
	}

	multiplier := computeMultiplier(req.ViewCount, account.AvgViewCount)

	var (
		titlePtr      *string
		urlPtr        *string
		transcriptPtr *string
	)
	if req.Title != "" {
		titlePtr = &req.Title
	}
	if req.URL != "" {
		urlPtr = &req.URL
	}
	if req.RawTranscript != "" {
		transcriptPtr = &req.RawTranscript
	}

	reel := &models.OutlierReel{
		CompetitorAccountID:  req.CompetitorAccountID,
		PlatformID:           req.PlatformID,
		ExternalReelID:       req.ExternalReelID,
		Title:                titlePtr,
		URL:                  urlPtr,
		ViewCount:            req.ViewCount,
		LikeCount:            req.LikeCount,
		CommentCount:         req.CommentCount,
		ShareCount:           req.ShareCount,
		AccountAvgAtCapture:  account.AvgViewCount,
		Multiplier:           multiplier,
		RawTranscript:        transcriptPtr,
		PublishedAt:          req.PublishedAt,
		RelevanceStatus:      models.RelevanceStatusPending,
	}

	if err := s.outlierReelRepo.Create(reel); err != nil {
		return nil, err
	}
	return reel, nil
}

// UpdateAccountBaseline recalculates and persists the avg view count and 5X threshold
// for a competitor account. Called after each Scout scan.
func (s *intelligenceService) UpdateAccountBaseline(accountID int32, newAvgViewCount int64) error {
	if newAvgViewCount <= 0 {
		return errors.New("avg view count must be positive")
	}
	threshold := newAvgViewCount * 5
	return s.competitorRepo.UpdateAvgAndThreshold(accountID, newAvgViewCount, threshold)
}

// --- Relevance filter ---

func (s *intelligenceService) ListPendingRelevanceFilter(limit int) ([]models.OutlierReel, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.outlierReelRepo.FindPendingRelevanceFilter(limit)
}

func (s *intelligenceService) ApproveReel(id int32, reason string) error {
	return s.outlierReelRepo.UpdateRelevanceStatus(id, models.RelevanceStatusApproved, reason)
}

func (s *intelligenceService) RejectReel(id int32, reason string) error {
	if reason == "" {
		return errors.New("rejection reason is required — agents must document why the reel is not authentic for Nick")
	}
	return s.outlierReelRepo.UpdateRelevanceStatus(id, models.RelevanceStatusRejected, reason)
}

// --- Trend signals ---

// UpsertTrendSignal creates or updates a trend signal for a topic+platform pair.
// Only Emerging and Peaking stages are actionable — Declining signals are stored but not routed.
func (s *intelligenceService) UpsertTrendSignal(req UpsertTrendSignalRequest) (*models.TrendSignal, error) {
	if err := validateTrendStage(req.Stage); err != nil {
		return nil, err
	}

	existing, err := s.trendSignalRepo.FindByTopicAndPlatform(req.Topic, req.PlatformID)
	if err == nil && existing != nil {
		// Update existing signal
		existing.Volume30d = req.Volume30d
		existing.Volume7d = req.Volume7d
		existing.Volume24h = req.Volume24h
		existing.Stage = req.Stage
		existing.UrgencyRating = req.UrgencyRating
		existing.DetectedAt = time.Now()
		if err := s.trendSignalRepo.Update(existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	signal := &models.TrendSignal{
		Topic:         req.Topic,
		PlatformID:    req.PlatformID,
		Volume30d:     req.Volume30d,
		Volume7d:      req.Volume7d,
		Volume24h:     req.Volume24h,
		Stage:         req.Stage,
		UrgencyRating: req.UrgencyRating,
	}
	if err := s.trendSignalRepo.Create(signal); err != nil {
		return nil, err
	}
	return signal, nil
}

func (s *intelligenceService) ListActionableTrendSignals() ([]models.TrendSignal, error) {
	return s.trendSignalRepo.FindActionable()
}

func (s *intelligenceService) GetTrendSignal(id int32) (*models.TrendSignal, error) {
	return s.trendSignalRepo.FindByID(id)
}

// ---------------------------------------------------------------------------
// Private helpers — reusable pure functions
// ---------------------------------------------------------------------------

// computeMultiplier calculates view_count / avg_view_count safely.
func computeMultiplier(viewCount, avgViewCount int64) float64 {
	if avgViewCount == 0 {
		return 0
	}
	return float64(viewCount) / float64(avgViewCount)
}

// validateTrendStage enforces the closed list of valid stage values.
func validateTrendStage(stage string) error {
	valid := map[string]bool{
		models.TrendStageEmerging:  true,
		models.TrendStagePeaking:   true,
		models.TrendStageSaturated: true,
		models.TrendStageDeclining: true,
	}
	if !valid[stage] {
		return errors.New("invalid trend stage — must be one of: emerging, peaking, saturated, declining")
	}
	return nil
}
