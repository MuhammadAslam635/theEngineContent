package repositories

import (
	"backend-go/internal/models"

	"gorm.io/gorm"
)

type AgentSettingRepository interface {
	FindAll() ([]models.AgentSetting, error)
	FindByID(id int32) (*models.AgentSetting, error)
	FindByName(name string) (*models.AgentSetting, error)
	Create(agent *models.AgentSetting) error
	Update(agent *models.AgentSetting) error
	Delete(id int32) error
}

type agentSettingRepository struct {
	baseRepository
}

func NewAgentSettingRepository(db *gorm.DB) AgentSettingRepository {
	return &agentSettingRepository{baseRepository{db: db}}
}

func (r *agentSettingRepository) FindAll() ([]models.AgentSetting, error) {
	var agents []models.AgentSetting
	if err := r.db.Order("created_at DESC").Find(&agents).Error; err != nil {
		return nil, err
	}
	return agents, nil
}

func (r *agentSettingRepository) FindByID(id int32) (*models.AgentSetting, error) {
	var agent models.AgentSetting
	if err := r.db.First(&agent, id).Error; err != nil {
		return nil, err
	}
	return &agent, nil
}

func (r *agentSettingRepository) FindByName(name string) (*models.AgentSetting, error) {
	var agent models.AgentSetting
	if err := r.db.Where("agent_name = ?", name).First(&agent).Error; err != nil {
		return nil, err
	}
	return &agent, nil
}

func (r *agentSettingRepository) Create(agent *models.AgentSetting) error {
	return r.db.Create(agent).Error
}

func (r *agentSettingRepository) Update(agent *models.AgentSetting) error {
	return r.db.Save(agent).Error
}

func (r *agentSettingRepository) Delete(id int32) error {
	return r.db.Delete(&models.AgentSetting{}, id).Error
}
