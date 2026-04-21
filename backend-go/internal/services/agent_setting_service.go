package services

import (
	"encoding/json"
	"errors"

	"backend-go/internal/dto"
	"backend-go/internal/models"
	"backend-go/internal/repositories"
)

type AgentSettingService interface {
	GetAll() ([]models.AgentSetting, error)
	GetByID(id int32) (*models.AgentSetting, error)
	GetByName(name string) (*models.AgentSetting, error)
	Create(req dto.CreateAgentSettingRequest) (*models.AgentSetting, error)
	Update(id int32, req dto.UpdateAgentSettingRequest) (*models.AgentSetting, error)
	Delete(id int32) error
}

type agentSettingService struct {
	repo repositories.AgentSettingRepository
}

func NewAgentSettingService(repo repositories.AgentSettingRepository) AgentSettingService {
	return &agentSettingService{repo: repo}
}

func (s *agentSettingService) GetAll() ([]models.AgentSetting, error) {
	return s.repo.FindAll()
}

func (s *agentSettingService) GetByID(id int32) (*models.AgentSetting, error) {
	agent, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("agent setting not found")
	}
	return agent, nil
}

func (s *agentSettingService) GetByName(name string) (*models.AgentSetting, error) {
	agent, err := s.repo.FindByName(name)
	if err != nil {
		return nil, errors.New("agent setting not found")
	}
	return agent, nil
}

func (s *agentSettingService) Create(req dto.CreateAgentSettingRequest) (*models.AgentSetting, error) {
	variables := req.Variables
	if variables == nil {
		variables = json.RawMessage("[]")
	}

	temperature := float32(0.7)
	if req.Temperature != nil {
		temperature = *req.Temperature
	}

	maxTokens := 2048
	if req.MaxTokens != nil {
		maxTokens = *req.MaxTokens
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	agent := &models.AgentSetting{
		AgentName:    req.AgentName,
		Supervisor:   req.Supervisor,
		Provider:     req.Provider,
		ModelName:    req.ModelName,
		Prompt:       req.Prompt,
		Variables:    variables,
		OutputType:   req.OutputType,
		OutputSchema: req.OutputSchema,
		Temperature:  temperature,
		MaxTokens:    maxTokens,
		IsActive:     isActive,
	}

	if err := s.repo.Create(agent); err != nil {
		return nil, err
	}
	return agent, nil
}

func (s *agentSettingService) Update(id int32, req dto.UpdateAgentSettingRequest) (*models.AgentSetting, error) {
	agent, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("agent setting not found")
	}

	if req.AgentName != nil {
		agent.AgentName = *req.AgentName
	}
	if req.Supervisor != nil {
		agent.Supervisor = req.Supervisor
	}
	if req.Provider != nil {
		agent.Provider = *req.Provider
	}
	if req.ModelName != nil {
		agent.ModelName = *req.ModelName
	}
	if req.Prompt != nil {
		agent.Prompt = req.Prompt
	}
	if req.Variables != nil {
		agent.Variables = req.Variables
	}
	if req.OutputType != nil {
		agent.OutputType = *req.OutputType
	}
	if req.OutputSchema != nil {
		agent.OutputSchema = req.OutputSchema
	}
	if req.Temperature != nil {
		agent.Temperature = *req.Temperature
	}
	if req.MaxTokens != nil {
		agent.MaxTokens = *req.MaxTokens
	}
	if req.IsActive != nil {
		agent.IsActive = *req.IsActive
	}

	if err := s.repo.Update(agent); err != nil {
		return nil, err
	}
	return agent, nil
}

func (s *agentSettingService) Delete(id int32) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("agent setting not found")
	}
	return s.repo.Delete(id)
}
