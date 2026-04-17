package services

import (
	"errors"

	"backend-go/internal/models"
	"backend-go/internal/repositories"
)

// ---------------------------------------------------------------------------
// ContentLibraryService
// Manages the closed-list libraries agents select from:
// hooks, angles, personas, Nick's credentials, and the approved scripts storehouse.
// Agents NEVER invent options — they select from what this service provides.
// ---------------------------------------------------------------------------

type ContentLibraryService interface {
	// Hooks
	GetHook(id int32) (*models.HookLibrary, error)
	ListTopHooks(limit int) ([]models.HookLibrary, error)
	ListHooksByTrigger(trigger string) ([]models.HookLibrary, error)
	RecordHookUsage(id int32, viewCount float64) error

	// Angles — seeded with 11, expanded from real outlier analysis
	GetAngle(id int32) (*models.AngleLibrary, error)
	GetAngleByKey(key string) (*models.AngleLibrary, error)
	ListAllAngles() ([]models.AngleLibrary, error)
	ListTopAngles(limit int) ([]models.AngleLibrary, error)
	RecordAngleUsage(id int32, viewCount float64) error

	// Personas — closed list of four
	GetPersona(id int32) (*models.Persona, error)
	GetPersonaByKey(key string) (*models.Persona, error)
	ListActivePersonas() ([]models.Persona, error)

	// Nick's verified credentials — agents pull from here only
	GetCredential(id int32) (*models.NickCredential, error)
	GetCredentialByKey(key string) (*models.NickCredential, error)
	ListCredentialsByCategory(category string) ([]models.NickCredential, error)
	ListAllActiveCredentials() ([]models.NickCredential, error)
	AddCredential(req AddCredentialRequest) (*models.NickCredential, error)

	// Approved scripts storehouse
	GetScript(id int32) (*models.ApprovedScript, error)
	ListActiveScripts(contentType string) ([]models.ApprovedScript, error)
	FindCandidatesForBrief(contentType string, personaID, angleID *int32) ([]models.ApprovedScript, error)
	AddScript(req AddScriptRequest) (*models.ApprovedScript, error)
	PromoteWriterAgentScript(req AddScriptRequest) (*models.ApprovedScript, error)
	RecordScriptUsage(id int32, viewCount float64) error
	ScriptstoreHealth() (ScriptstoreHealthReport, error)
}

// ---------------------------------------------------------------------------
// Request / response types
// ---------------------------------------------------------------------------

type AddCredentialRequest struct {
	CredentialKey string
	Category      string
	DisplayValue  string
	RawValue      *float64
	Unit          string
	Context       string
}

type AddScriptRequest struct {
	OutlierReelID         *int32
	Title                 string
	ScriptBody            string
	ContentType           string
	Source                string
	PersonaID             *int32
	AngleID               *int32
	HookID                *int32
	TopicTags             []string
	WordCount             int
	EstimatedDurationSecs int
	VoiceDNAScore         float64
	ValidationScore       float64
	ValidationNotes       string
}

// ScriptstoreHealthReport — dashboard metric ensuring the database never runs empty.
type ScriptstoreHealthReport struct {
	ShortFormCount int64
	LongFormCount  int64
	IsHealthy      bool   // false if either count drops to zero
	Warning        string
}

// ---------------------------------------------------------------------------
// Implementation
// ---------------------------------------------------------------------------

type contentLibraryService struct {
	hookRepo       repositories.HookLibraryRepository
	angleRepo      repositories.AngleLibraryRepository
	personaRepo    repositories.PersonaRepository
	credentialRepo repositories.NickCredentialRepository
	scriptRepo     repositories.ApprovedScriptRepository
}

func NewContentLibraryService(
	hookRepo repositories.HookLibraryRepository,
	angleRepo repositories.AngleLibraryRepository,
	personaRepo repositories.PersonaRepository,
	credentialRepo repositories.NickCredentialRepository,
	scriptRepo repositories.ApprovedScriptRepository,
) ContentLibraryService {
	return &contentLibraryService{
		hookRepo:       hookRepo,
		angleRepo:      angleRepo,
		personaRepo:    personaRepo,
		credentialRepo: credentialRepo,
		scriptRepo:     scriptRepo,
	}
}

// --- Hooks ---

func (s *contentLibraryService) GetHook(id int32) (*models.HookLibrary, error) {
	return s.hookRepo.FindByID(id)
}

func (s *contentLibraryService) ListTopHooks(limit int) ([]models.HookLibrary, error) {
	return s.hookRepo.FindTopPerforming(limit)
}

func (s *contentLibraryService) ListHooksByTrigger(trigger string) ([]models.HookLibrary, error) {
	if err := validateEmotionalTrigger(trigger); err != nil {
		return nil, err
	}
	return s.hookRepo.FindByEmotionalTrigger(trigger)
}

func (s *contentLibraryService) RecordHookUsage(id int32, viewCount float64) error {
	return s.hookRepo.IncrementUsage(id, viewCount)
}

// --- Angles ---

func (s *contentLibraryService) GetAngle(id int32) (*models.AngleLibrary, error) {
	return s.angleRepo.FindByID(id)
}

func (s *contentLibraryService) GetAngleByKey(key string) (*models.AngleLibrary, error) {
	return s.angleRepo.FindByKey(key)
}

func (s *contentLibraryService) ListAllAngles() ([]models.AngleLibrary, error) {
	return s.angleRepo.FindAll()
}

func (s *contentLibraryService) ListTopAngles(limit int) ([]models.AngleLibrary, error) {
	return s.angleRepo.FindTopPerforming(limit)
}

func (s *contentLibraryService) RecordAngleUsage(id int32, viewCount float64) error {
	return s.angleRepo.IncrementUsage(id, viewCount)
}

// --- Personas ---

func (s *contentLibraryService) GetPersona(id int32) (*models.Persona, error) {
	return s.personaRepo.FindByID(id)
}

func (s *contentLibraryService) GetPersonaByKey(key string) (*models.Persona, error) {
	if err := validatePersonaKey(key); err != nil {
		return nil, err
	}
	return s.personaRepo.FindByKey(key)
}

func (s *contentLibraryService) ListActivePersonas() ([]models.Persona, error) {
	return s.personaRepo.FindAllActive()
}

// --- Nick's verified credentials ---

func (s *contentLibraryService) GetCredential(id int32) (*models.NickCredential, error) {
	return s.credentialRepo.FindByID(id)
}

func (s *contentLibraryService) GetCredentialByKey(key string) (*models.NickCredential, error) {
	return s.credentialRepo.FindByKey(key)
}

func (s *contentLibraryService) ListCredentialsByCategory(category string) ([]models.NickCredential, error) {
	return s.credentialRepo.FindByCategory(category)
}

func (s *contentLibraryService) ListAllActiveCredentials() ([]models.NickCredential, error) {
	return s.credentialRepo.FindAllActive()
}

func (s *contentLibraryService) AddCredential(req AddCredentialRequest) (*models.NickCredential, error) {
	if req.CredentialKey == "" {
		return nil, errors.New("credential_key is required")
	}
	if req.DisplayValue == "" {
		return nil, errors.New("display_value is required — must be the exact phrasing used in scripts")
	}
	var unitPtr *string
	if req.Unit != "" {
		unitPtr = &req.Unit
	}
	var contextPtr *string
	if req.Context != "" {
		contextPtr = &req.Context
	}
	cred := &models.NickCredential{
		CredentialKey: req.CredentialKey,
		Category:      req.Category,
		DisplayValue:  req.DisplayValue,
		RawValue:      req.RawValue,
		Unit:          unitPtr,
		Context:       contextPtr,
	}
	if err := s.credentialRepo.Create(cred); err != nil {
		return nil, err
	}
	return cred, nil
}

// --- Approved scripts storehouse ---

func (s *contentLibraryService) GetScript(id int32) (*models.ApprovedScript, error) {
	return s.scriptRepo.FindByID(id)
}

func (s *contentLibraryService) ListActiveScripts(contentType string) ([]models.ApprovedScript, error) {
	return s.scriptRepo.FindAllActive(contentType)
}

func (s *contentLibraryService) FindCandidatesForBrief(contentType string, personaID, angleID *int32) ([]models.ApprovedScript, error) {
	if err := validateContentType(contentType); err != nil {
		return nil, err
	}
	return s.scriptRepo.FindBestForBrief(contentType, personaID, angleID)
}

func (s *contentLibraryService) AddScript(req AddScriptRequest) (*models.ApprovedScript, error) {
	return s.createScript(req)
}

func (s *contentLibraryService) PromoteWriterAgentScript(req AddScriptRequest) (*models.ApprovedScript, error) {
	req.Source = models.ScriptSourceWriterAgent
	script, err := s.createScript(req)
	if err != nil {
		return nil, err
	}
	return script, s.scriptRepo.PromoteFromWriterAgent(script)
}

func (s *contentLibraryService) RecordScriptUsage(id int32, viewCount float64) error {
	return s.scriptRepo.IncrementUsage(id, viewCount)
}

// ScriptstoreHealth — called by dashboard and monitoring; warns when either pool runs empty.
func (s *contentLibraryService) ScriptstoreHealth() (ScriptstoreHealthReport, error) {
	shortCount, err := s.scriptRepo.CountActive(models.ContentTypeShortForm)
	if err != nil {
		return ScriptstoreHealthReport{}, err
	}
	longCount, err := s.scriptRepo.CountActive(models.ContentTypeLongForm)
	if err != nil {
		return ScriptstoreHealthReport{}, err
	}

	report := ScriptstoreHealthReport{
		ShortFormCount: shortCount,
		LongFormCount:  longCount,
		IsHealthy:      shortCount > 0 && longCount > 0,
	}
	if shortCount == 0 {
		report.Warning = "CRITICAL: short-form script storehouse is empty — pipeline will stall"
	} else if longCount == 0 {
		report.Warning = "CRITICAL: long-form script storehouse is empty — long-form pipeline will stall"
	}
	return report, nil
}

// ---------------------------------------------------------------------------
// Private helpers
// ---------------------------------------------------------------------------

func (s *contentLibraryService) createScript(req AddScriptRequest) (*models.ApprovedScript, error) {
	if req.Title == "" {
		return nil, errors.New("script title is required")
	}
	if req.ScriptBody == "" {
		return nil, errors.New("script body is required")
	}
	if err := validateContentType(req.ContentType); err != nil {
		return nil, err
	}
	// VoiceDNA score must meet threshold to be added to approved storehouse
	if req.VoiceDNAScore < models.ScriptPassThreshold {
		return nil, errors.New("script VoiceDNA score must be >= 90 to enter the approved storehouse")
	}

	var notesPtr *string
	if req.ValidationNotes != "" {
		notesPtr = &req.ValidationNotes
	}

	script := &models.ApprovedScript{
		OutlierReelID:         req.OutlierReelID,
		Title:                 req.Title,
		ScriptBody:            req.ScriptBody,
		ContentType:           req.ContentType,
		Source:                req.Source,
		PersonaID:             req.PersonaID,
		AngleID:               req.AngleID,
		HookID:                req.HookID,
		TopicTags:             req.TopicTags,
		WordCount:             req.WordCount,
		EstimatedDurationSecs: req.EstimatedDurationSecs,
		VoiceDNAScore:         req.VoiceDNAScore,
		ValidationScore:       req.ValidationScore,
		ValidationNotes:       notesPtr,
	}
	if err := s.scriptRepo.Create(script); err != nil {
		return nil, err
	}
	return script, nil
}

// --- Closed-list validators — reused across services ---

func validateEmotionalTrigger(trigger string) error {
	valid := map[string]bool{
		models.EmotionalTriggerFrustration: true,
		models.EmotionalTriggerCuriosity:   true,
		models.EmotionalTriggerFear:        true,
		models.EmotionalTriggerAspiration:  true,
	}
	if !valid[trigger] {
		return errors.New("invalid emotional trigger — must be: frustration, curiosity, fear, or aspiration")
	}
	return nil
}

func validatePersonaKey(key string) error {
	valid := map[string]bool{
		models.PersonaKeyW2Escapee:            true,
		models.PersonaKeyStuckInvestor:        true,
		models.PersonaKeyAspiringEntrepreneur: true,
		models.PersonaKeyIndustrySwitcher:     true,
	}
	if !valid[key] {
		return errors.New("invalid persona key — must be one of the four approved personas")
	}
	return nil
}

func validateContentType(contentType string) error {
	valid := map[string]bool{
		models.ContentTypeShortForm: true,
		models.ContentTypeLongForm:  true,
	}
	if !valid[contentType] {
		return errors.New("invalid content type — must be short_form or long_form")
	}
	return nil
}
