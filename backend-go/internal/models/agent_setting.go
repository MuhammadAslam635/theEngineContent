package models

import (
	"encoding/json"
	"time"
)

// AgentVariable describes a placeholder variable that can be injected into the prompt.
type AgentVariable struct {
	Key         string `json:"key"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// AgentSetting — dynamic agent configuration stored in DB.
type AgentSetting struct {
	ID           int32            `gorm:"primaryKey;autoIncrement" json:"id"`
	AgentName    string           `gorm:"size:255;not null" json:"agent_name"`
	Supervisor   *string          `gorm:"size:255" json:"supervisor,omitempty"`
	Provider     string           `gorm:"size:100;not null" json:"provider"`
	ModelName    string           `gorm:"size:255;not null" json:"model_name"`
	Prompt       *string          `json:"prompt,omitempty"`
	Variables    json.RawMessage  `gorm:"type:jsonb;default:'[]'" json:"variables"`
	OutputType   string           `gorm:"size:10;default:text" json:"output_type"`   // text | json
	OutputSchema json.RawMessage  `gorm:"type:jsonb" json:"output_schema,omitempty"` // JSON schema when output_type = json
	Temperature  float32          `gorm:"default:0.7" json:"temperature"`
	MaxTokens    int              `gorm:"default:2048" json:"max_tokens"`
	IsActive     bool             `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

// OutputType constants
const (
	AgentOutputText = "text"
	AgentOutputJSON = "json"
)
