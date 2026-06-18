package model

import (
	"encoding/json"
	"time"
)

type Agent struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Icon        string    `json:"icon" db:"icon"`
	Color       string    `json:"color" db:"color"`
	Category    string    `json:"category" db:"category"`
	ShortDesc   string    `json:"short_desc" db:"short_desc"`
	FullDesc    string    `json:"full_desc" db:"full_desc"`
	Tags        []string  `json:"tags" db:"tags"`
	Cost        int       `json:"cost" db:"cost"`
	Status      string    `json:"status" db:"status"`
	Prompt      string    `json:"prompt" db:"prompt"`
	Temperature float64   `json:"temperature" db:"temperature"`
	MaxTokens   int       `json:"max_tokens" db:"max_tokens"`
	ModelName   string    `json:"model_name" db:"model_name"`
	BaseURL     string    `json:"base_url" db:"base_url"`
	APIKey      string    `json:"-" db:"api_key"`
	Rating      float64   `json:"rating" db:"rating"`
	Calls       int64     `json:"calls" db:"calls"`
	Featured    bool      `json:"featured" db:"featured"`
	Speed       string    `json:"speed" db:"speed"`
	Precision   string    `json:"precision" db:"precision"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type AgentTool struct {
	ID          int64           `json:"id" db:"id"`
	AgentID     int64           `json:"agent_id" db:"agent_id"`
	Name        string          `json:"name" db:"name"`
	Description string          `json:"description" db:"description"`
	Type        string          `json:"type" db:"type"`
	InputSchema json.RawMessage `json:"input_schema" db:"input_schema"`
	Config      json.RawMessage `json:"-" db:"config"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

const (
	AgentStatusActive   = "active"
	AgentStatusInactive = "inactive"

	ToolTypeCoze = "coze"
	ToolTypeN8N  = "n8n"
)
