package model

import "time"

type Message struct {
	ID             int64     `json:"id" db:"id"`
	ConversationID int64     `json:"conversation_id" db:"conversation_id"`
	Role           string    `json:"role" db:"role"`
	Content        string    `json:"content" db:"content"`
	ToolCalls      string    `json:"tool_calls,omitempty" db:"tool_calls"`
	ToolCallID     string    `json:"tool_call_id,omitempty" db:"tool_call_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
)
