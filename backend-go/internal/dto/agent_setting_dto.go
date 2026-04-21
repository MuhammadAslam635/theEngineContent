package dto

import "encoding/json"

// CreateAgentSettingRequest is the payload for creating a new agent setting.
type CreateAgentSettingRequest struct {
	AgentName    string          `json:"agent_name" binding:"required"`
	Supervisor   *string         `json:"supervisor"`
	Provider     string          `json:"provider" binding:"required"`
	ModelName    string          `json:"model_name" binding:"required"`
	Prompt       *string         `json:"prompt"`
	Variables    json.RawMessage `json:"variables"`
	OutputType   string          `json:"output_type" binding:"required,oneof=text json"`
	OutputSchema json.RawMessage `json:"output_schema"`
	Temperature  *float32        `json:"temperature"`
	MaxTokens    *int            `json:"max_tokens"`
	IsActive     *bool           `json:"is_active"`
}

// UpdateAgentSettingRequest is the payload for updating an existing agent setting.
type UpdateAgentSettingRequest struct {
	AgentName    *string         `json:"agent_name"`
	Supervisor   *string         `json:"supervisor"`
	Provider     *string         `json:"provider"`
	ModelName    *string         `json:"model_name"`
	Prompt       *string         `json:"prompt"`
	Variables    json.RawMessage `json:"variables"`
	OutputType   *string         `json:"output_type" binding:"omitempty,oneof=text json"`
	OutputSchema json.RawMessage `json:"output_schema"`
	Temperature  *float32        `json:"temperature"`
	MaxTokens    *int            `json:"max_tokens"`
	IsActive     *bool           `json:"is_active"`
}
