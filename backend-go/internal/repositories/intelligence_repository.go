package repositories

import (
	"backend-go/internal/models"

	"gorm.io/gorm"
)

// ===========================================================================
// PlatformRepository
// ===========================================================================

type PlatformRepository interface {
	FindAll() ([]models.Platform, error)
	FindByID(id int32) (*models.Platform, error)
	FindByName(name string) (*models.Platform, error)
	FindAllActive() ([]models.Platform, error)
	Update(platform *models.Platform) error
}

type platformRepository struct {
	baseRepository
}

func NewPlatformRepository(db *gorm.DB) PlatformRepository {
	return &platformRepository{baseRepository{db: db}}
}

func (r *platformRepository) FindAll() ([]models.Platform, error) {
	var platforms []models.Platform
	err := r.db.Find(&platforms).Error
	return platforms, err
}

func (r *platformRepository) FindByID(id int32) (*models.Platform, error) {
	var platform models.Platform
	err := r.db.First(&platform, id).Error
	return &platform, err
}

func (r *platformRepository) FindByName(name string) (*models.Platform, error) {
	var platform models.Platform
	err := r.db.Where("name = ?", name).First(&platform).Error
	return &platform, err
}

func (r *platformRepository) FindAllActive() ([]models.Platform, error) {
	var platforms []models.Platform
	err := applyActiveFilter(r.db).Find(&platforms).Error
	return platforms, err
}

func (r *platformRepository) Update(platform *models.Platform) error {
	return r.db.Save(platform).Error
}

// ===========================================================================
// CompetitorAccountRepository
// ===========================================================================

type CompetitorAccountRepository interface {
	Create(account *models.CompetitorAccount) error
	FindByID(id int32) (*models.CompetitorAccount, error)
	FindAllActive() ([]models.CompetitorAccount, error)
	FindByPlatform(platformID int32) ([]models.CompetitorAccount, error)
	UpdateAvgAndThreshold(id int32, avgViewCount, threshold int64) error
	Update(account *models.CompetitorAccount) error
	Delete(id int32) error
}

type competitorAccountRepository struct {
	baseRepository
}

func NewCompetitorAccountRepository(db *gorm.DB) CompetitorAccountRepository {
	return &competitorAccountRepository{baseRepository{db: db}}
}

func (r *competitorAccountRepository) Create(account *models.CompetitorAccount) error {
	return r.db.Create(account).Error
}

func (r *competitorAccountRepository) FindByID(id int32) (*models.CompetitorAccount, error) {
	var account models.CompetitorAccount
	err := r.db.Preload("Platform").First(&account, id).Error
	return &account, err
}

func (r *competitorAccountRepository) FindAllActive() ([]models.CompetitorAccount, error) {
	var accounts []models.CompetitorAccount
	err := applyActiveFilter(r.db).Preload("Platform").Find(&accounts).Error
	return accounts, err
}

func (r *competitorAccountRepository) FindByPlatform(platformID int32) ([]models.CompetitorAccount, error) {
	var accounts []models.CompetitorAccount
	err := applyActiveFilter(r.db).
		Where("platform_id = ?", platformID).
		Find(&accounts).Error
	return accounts, err
}

func (r *competitorAccountRepository) UpdateAvgAndThreshold(id int32, avgViewCount, threshold int64) error {
	return r.db.Model(&models.CompetitorAccount{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"avg_view_count":    avgViewCount,
			"outlier_threshold": threshold,
		}).Error
}

func (r *competitorAccountRepository) Update(account *models.CompetitorAccount) error {
	return r.db.Save(account).Error
}

func (r *competitorAccountRepository) Delete(id int32) error {
	return r.db.Model(&models.CompetitorAccount{}).Where("id = ?", id).Update("is_active", false).Error
}

// ===========================================================================
// OutlierReelRepository
// ===========================================================================

type OutlierReelRepository interface {
	Create(reel *models.OutlierReel) error
	FindByID(id int32) (*models.OutlierReel, error)
	FindByExternalID(platformID int32, externalReelID string) (*models.OutlierReel, error)
	FindPendingRelevanceFilter(limit int) ([]models.OutlierReel, error)
	FindApproved() ([]models.OutlierReel, error)
	UpdateRelevanceStatus(id int32, status, reason string) error
	Update(reel *models.OutlierReel) error
}

type outlierReelRepository struct {
	baseRepository
}

func NewOutlierReelRepository(db *gorm.DB) OutlierReelRepository {
	return &outlierReelRepository{baseRepository{db: db}}
}

func (r *outlierReelRepository) Create(reel *models.OutlierReel) error {
	return r.db.Create(reel).Error
}

func (r *outlierReelRepository) FindByID(id int32) (*models.OutlierReel, error) {
	var reel models.OutlierReel
	err := r.db.Preload("CompetitorAccount").Preload("Platform").First(&reel, id).Error
	return &reel, err
}

func (r *outlierReelRepository) FindByExternalID(platformID int32, externalReelID string) (*models.OutlierReel, error) {
	var reel models.OutlierReel
	err := r.db.Where("platform_id = ? AND external_reel_id = ?", platformID, externalReelID).
		First(&reel).Error
	return &reel, err
}

func (r *outlierReelRepository) FindPendingRelevanceFilter(limit int) ([]models.OutlierReel, error) {
	var reels []models.OutlierReel
	err := r.db.Where("relevance_status = ?", models.RelevanceStatusPending).
		Order("detected_at DESC").
		Limit(limit).
		Find(&reels).Error
	return reels, err
}

func (r *outlierReelRepository) FindApproved() ([]models.OutlierReel, error) {
	var reels []models.OutlierReel
	err := r.db.Where("relevance_status = ?", models.RelevanceStatusApproved).
		Order("detected_at DESC").
		Find(&reels).Error
	return reels, err
}

func (r *outlierReelRepository) UpdateRelevanceStatus(id int32, status, reason string) error {
	return r.db.Model(&models.OutlierReel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"relevance_status":     status,
			"relevance_reason":     reason,
			"relevance_checked_at": gorm.Expr("NOW()"),
		}).Error
}

func (r *outlierReelRepository) Update(reel *models.OutlierReel) error {
	return r.db.Save(reel).Error
}

// ===========================================================================
// TrendSignalRepository
// ===========================================================================

type TrendSignalRepository interface {
	Create(signal *models.TrendSignal) error
	FindByID(id int32) (*models.TrendSignal, error)
	FindByStage(stage string) ([]models.TrendSignal, error)
	FindByTopicAndPlatform(topic string, platformID int32) (*models.TrendSignal, error)
	FindActionable() ([]models.TrendSignal, error) // emerging + peaking only
	Update(signal *models.TrendSignal) error
}

type trendSignalRepository struct {
	baseRepository
}

func NewTrendSignalRepository(db *gorm.DB) TrendSignalRepository {
	return &trendSignalRepository{baseRepository{db: db}}
}

func (r *trendSignalRepository) Create(signal *models.TrendSignal) error {
	return r.db.Create(signal).Error
}

func (r *trendSignalRepository) FindByID(id int32) (*models.TrendSignal, error) {
	var signal models.TrendSignal
	err := r.db.Preload("Platform").First(&signal, id).Error
	return &signal, err
}

func (r *trendSignalRepository) FindByStage(stage string) ([]models.TrendSignal, error) {
	var signals []models.TrendSignal
	err := r.db.Where("stage = ?", stage).
		Order("urgency_rating DESC, detected_at DESC").
		Find(&signals).Error
	return signals, err
}

func (r *trendSignalRepository) FindByTopicAndPlatform(topic string, platformID int32) (*models.TrendSignal, error) {
	var signal models.TrendSignal
	err := r.db.Where("topic = ? AND platform_id = ?", topic, platformID).
		Order("detected_at DESC").
		First(&signal).Error
	return &signal, err
}

func (r *trendSignalRepository) FindActionable() ([]models.TrendSignal, error) {
	var signals []models.TrendSignal
	err := r.db.Where("stage IN ?", []string{models.TrendStageEmerging, models.TrendStagePeaking}).
		Order("urgency_rating DESC").
		Find(&signals).Error
	return signals, err
}

func (r *trendSignalRepository) Update(signal *models.TrendSignal) error {
	return r.db.Save(signal).Error
}
