package model

import "time"

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
	Engine      string    `json:"engine" db:"engine"`
	Status      string    `json:"status" db:"status"`
	Prompt      string    `json:"prompt" db:"prompt"`
	Temperature float64   `json:"temperature" db:"temperature"`
	MaxTokens   int       `json:"max_tokens" db:"max_tokens"`
	Rating      float64   `json:"rating" db:"rating"`
	Calls       int64     `json:"calls" db:"calls"`
	Featured    bool      `json:"featured" db:"featured"`
	Speed       string    `json:"speed" db:"speed"`
	Precision   string    `json:"precision" db:"precision"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CozeWorkflow struct {
	AgentID     int64  `json:"agent_id" db:"agent_id"`
	WorkflowID  string `json:"workflow_id" db:"workflow_id"`
	APIKey      string `json:"-" db:"api_key"`
	Region      string `json:"region" db:"region"`
	InputField  string `json:"input_field" db:"input_field"`
	OutputField string `json:"output_field" db:"output_field"`
}

type N8NWorkflow struct {
	AgentID     int64  `json:"agent_id" db:"agent_id"`
	WebhookURL  string `json:"webhook_url" db:"webhook_url"`
	AuthType    string `json:"auth_type" db:"auth_type"`
	AuthToken   string `json:"-" db:"auth_token"`
	Timeout     int    `json:"timeout" db:"timeout"`
	PayloadTmpl string `json:"payload_tmpl" db:"payload_tmpl"`
}

const (
	AgentStatusActive   = "active"
	AgentStatusInactive = "inactive"

	EngineCoze = "coze"
	EngineN8N  = "n8n"
)
