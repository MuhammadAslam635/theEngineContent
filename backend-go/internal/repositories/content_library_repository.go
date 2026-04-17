package repositories

import (
	"backend-go/internal/models"

	"gorm.io/gorm"
)

// ===========================================================================
// HookLibraryRepository
// ===========================================================================

type HookLibraryRepository interface {
	Create(hook *models.HookLibrary) error
	FindByID(id int32) (*models.HookLibrary, error)
	FindByEmotionalTrigger(trigger string) ([]models.HookLibrary, error)
	FindTopPerforming(limit int) ([]models.HookLibrary, error)
	IncrementUsage(id int32, newAvgPerformance float64) error
	Update(hook *models.HookLibrary) error
}

type hookLibraryRepository struct {
	baseRepository
}

func NewHookLibraryRepository(db *gorm.DB) HookLibraryRepository {
	return &hookLibraryRepository{baseRepository{db: db}}
}

func (r *hookLibraryRepository) Create(hook *models.HookLibrary) error {
	return r.db.Create(hook).Error
}

func (r *hookLibraryRepository) FindByID(id int32) (*models.HookLibrary, error) {
	var hook models.HookLibrary
	err := r.db.Preload("OutlierReel").First(&hook, id).Error
	return &hook, err
}

func (r *hookLibraryRepository) FindByEmotionalTrigger(trigger string) ([]models.HookLibrary, error) {
	var hooks []models.HookLibrary
	err := r.db.Where("emotional_trigger = ?", trigger).
		Order("avg_performance DESC").
		Find(&hooks).Error
	return hooks, err
}

func (r *hookLibraryRepository) FindTopPerforming(limit int) ([]models.HookLibrary, error) {
	var hooks []models.HookLibrary
	err := r.db.Order("avg_performance DESC").Limit(limit).Find(&hooks).Error
	return hooks, err
}

func (r *hookLibraryRepository) IncrementUsage(id int32, newAvgPerformance float64) error {
	return r.db.Model(&models.HookLibrary{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"usage_count":     gorm.Expr("usage_count + 1"),
			"avg_performance": newAvgPerformance,
		}).Error
}

func (r *hookLibraryRepository) Update(hook *models.HookLibrary) error {
	return r.db.Save(hook).Error
}

// ===========================================================================
// AngleLibraryRepository
// ===========================================================================

type AngleLibraryRepository interface {
	FindAll() ([]models.AngleLibrary, error)
	FindByID(id int32) (*models.AngleLibrary, error)
	FindByKey(angleKey string) (*models.AngleLibrary, error)
	FindTopPerforming(limit int) ([]models.AngleLibrary, error)
	IncrementUsage(id int32, newAvgPerformance float64) error
	Create(angle *models.AngleLibrary) error
	Update(angle *models.AngleLibrary) error
}

type angleLibraryRepository struct {
	baseRepository
}

func NewAngleLibraryRepository(db *gorm.DB) AngleLibraryRepository {
	return &angleLibraryRepository{baseRepository{db: db}}
}

func (r *angleLibraryRepository) FindAll() ([]models.AngleLibrary, error) {
	var angles []models.AngleLibrary
	err := r.db.Order("avg_performance DESC").Find(&angles).Error
	return angles, err
}

func (r *angleLibraryRepository) FindByID(id int32) (*models.AngleLibrary, error) {
	var angle models.AngleLibrary
	err := r.db.First(&angle, id).Error
	return &angle, err
}

func (r *angleLibraryRepository) FindByKey(angleKey string) (*models.AngleLibrary, error) {
	var angle models.AngleLibrary
	err := r.db.Where("angle_key = ?", angleKey).First(&angle).Error
	return &angle, err
}

func (r *angleLibraryRepository) FindTopPerforming(limit int) ([]models.AngleLibrary, error) {
	var angles []models.AngleLibrary
	err := r.db.Order("avg_performance DESC").Limit(limit).Find(&angles).Error
	return angles, err
}

func (r *angleLibraryRepository) IncrementUsage(id int32, newAvgPerformance float64) error {
	return r.db.Model(&models.AngleLibrary{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"usage_count":     gorm.Expr("usage_count + 1"),
			"avg_performance": newAvgPerformance,
		}).Error
}

func (r *angleLibraryRepository) Create(angle *models.AngleLibrary) error {
	return r.db.Create(angle).Error
}

func (r *angleLibraryRepository) Update(angle *models.AngleLibrary) error {
	return r.db.Save(angle).Error
}

// ===========================================================================
// PersonaRepository
// ===========================================================================

type PersonaRepository interface {
	FindAll() ([]models.Persona, error)
	FindByID(id int32) (*models.Persona, error)
	FindByKey(personaKey string) (*models.Persona, error)
	FindAllActive() ([]models.Persona, error)
}

type personaRepository struct {
	baseRepository
}

func NewPersonaRepository(db *gorm.DB) PersonaRepository {
	return &personaRepository{baseRepository{db: db}}
}

func (r *personaRepository) FindAll() ([]models.Persona, error) {
	var personas []models.Persona
	err := r.db.Find(&personas).Error
	return personas, err
}

func (r *personaRepository) FindByID(id int32) (*models.Persona, error) {
	var persona models.Persona
	err := r.db.First(&persona, id).Error
	return &persona, err
}

func (r *personaRepository) FindByKey(personaKey string) (*models.Persona, error) {
	var persona models.Persona
	err := r.db.Where("persona_key = ?", personaKey).First(&persona).Error
	return &persona, err
}

func (r *personaRepository) FindAllActive() ([]models.Persona, error) {
	var personas []models.Persona
	err := applyActiveFilter(r.db).Find(&personas).Error
	return personas, err
}

// ===========================================================================
// NickCredentialRepository
// ===========================================================================

type NickCredentialRepository interface {
	FindAll() ([]models.NickCredential, error)
	FindByID(id int32) (*models.NickCredential, error)
	FindByKey(credentialKey string) (*models.NickCredential, error)
	FindByCategory(category string) ([]models.NickCredential, error)
	FindAllActive() ([]models.NickCredential, error)
	Create(cred *models.NickCredential) error
	Update(cred *models.NickCredential) error
}

type nickCredentialRepository struct {
	baseRepository
}

func NewNickCredentialRepository(db *gorm.DB) NickCredentialRepository {
	return &nickCredentialRepository{baseRepository{db: db}}
}

func (r *nickCredentialRepository) FindAll() ([]models.NickCredential, error) {
	var creds []models.NickCredential
	err := r.db.Find(&creds).Error
	return creds, err
}

func (r *nickCredentialRepository) FindByID(id int32) (*models.NickCredential, error) {
	var cred models.NickCredential
	err := r.db.First(&cred, id).Error
	return &cred, err
}

func (r *nickCredentialRepository) FindByKey(credentialKey string) (*models.NickCredential, error) {
	var cred models.NickCredential
	err := r.db.Where("credential_key = ?", credentialKey).First(&cred).Error
	return &cred, err
}

func (r *nickCredentialRepository) FindByCategory(category string) ([]models.NickCredential, error) {
	var creds []models.NickCredential
	err := applyActiveFilter(r.db).Where("category = ?", category).Find(&creds).Error
	return creds, err
}

func (r *nickCredentialRepository) FindAllActive() ([]models.NickCredential, error) {
	var creds []models.NickCredential
	err := applyActiveFilter(r.db).Find(&creds).Error
	return creds, err
}

func (r *nickCredentialRepository) Create(cred *models.NickCredential) error {
	return r.db.Create(cred).Error
}

func (r *nickCredentialRepository) Update(cred *models.NickCredential) error {
	return r.db.Save(cred).Error
}

// ===========================================================================
// ApprovedScriptRepository
// ===========================================================================

type ApprovedScriptRepository interface {
	Create(script *models.ApprovedScript) error
	FindByID(id int32) (*models.ApprovedScript, error)
	FindAllActive(contentType string) ([]models.ApprovedScript, error)
	FindBestForBrief(contentType string, personaID, angleID *int32) ([]models.ApprovedScript, error)
	PromoteFromWriterAgent(script *models.ApprovedScript) error
	IncrementUsage(id int32, viewCount float64) error
	CountActive(contentType string) (int64, error)
	Update(script *models.ApprovedScript) error
}

type approvedScriptRepository struct {
	baseRepository
}

func NewApprovedScriptRepository(db *gorm.DB) ApprovedScriptRepository {
	return &approvedScriptRepository{baseRepository{db: db}}
}

func (r *approvedScriptRepository) Create(script *models.ApprovedScript) error {
	return r.db.Create(script).Error
}

func (r *approvedScriptRepository) FindByID(id int32) (*models.ApprovedScript, error) {
	var script models.ApprovedScript
	err := r.db.Preload("Persona").Preload("Angle").Preload("Hook").First(&script, id).Error
	return &script, err
}

func (r *approvedScriptRepository) FindAllActive(contentType string) ([]models.ApprovedScript, error) {
	var scripts []models.ApprovedScript
	q := applyActiveFilter(r.db)
	if contentType != "" {
		q = q.Where("content_type = ?", contentType)
	}
	err := q.Order("validation_score DESC").Find(&scripts).Error
	return scripts, err
}

// FindBestForBrief — returns candidate scripts ordered by validation_score for
// the Validation Agent to score. Filters by content_type; optionally narrows
// by persona and angle if provided.
func (r *approvedScriptRepository) FindBestForBrief(contentType string, personaID, angleID *int32) ([]models.ApprovedScript, error) {
	var scripts []models.ApprovedScript
	q := applyActiveFilter(r.db).Where("content_type = ?", contentType)
	if personaID != nil {
		q = q.Where("persona_id = ? OR persona_id IS NULL", *personaID)
	}
	if angleID != nil {
		q = q.Where("angle_id = ? OR angle_id IS NULL", *angleID)
	}
	err := q.Order("validation_score DESC").Find(&scripts).Error
	return scripts, err
}

// PromoteFromWriterAgent — adds a Writer Agent generated script to the approved database.
// Called automatically when a generated script exceeds the performance threshold after publishing.
func (r *approvedScriptRepository) PromoteFromWriterAgent(script *models.ApprovedScript) error {
	script.Source = models.ScriptSourceWriterAgent
	return r.db.Create(script).Error
}

func (r *approvedScriptRepository) IncrementUsage(id int32, viewCount float64) error {
	return r.db.Model(&models.ApprovedScript{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"times_used":          gorm.Expr("times_used + 1"),
			"last_used_at":        gorm.Expr("NOW()"),
			"avg_views_when_used": viewCount,
		}).Error
}

func (r *approvedScriptRepository) CountActive(contentType string) (int64, error) {
	q := applyActiveFilter(r.db).Model(&models.ApprovedScript{})
	if contentType != "" {
		q = q.Where("content_type = ?", contentType)
	}
	return countRows(q, &models.ApprovedScript{})
}

func (r *approvedScriptRepository) Update(script *models.ApprovedScript) error {
	return r.db.Save(script).Error
}
