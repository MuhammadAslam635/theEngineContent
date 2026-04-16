package models

import (
	"time"
)

type AuditLog struct {
	ID                uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Agent             string    `gorm:"not null" json:"agent"`
	InputQuery        string    `json:"input_query"`
	AgentPrompt       string    `json:"agent_prompt"`
	InputTokens       int       `gorm:"default:0" json:"input_tokens"`
	OutputResponse    string    `json:"output_response"`
	OutputUsageTokens int       `gorm:"default:0" json:"output_usage_tokens"`
	Error             string    `json:"error"`
	LineNumber        int       `json:"line_number"`
	UserID            int32     `gorm:"not null" json:"user_id"`
	CreatedAt         time.Time `json:"created_at"`
}
