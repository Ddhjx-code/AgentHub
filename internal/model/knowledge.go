package model

import "time"

type KnowledgeBase struct {
	ID               int64     `json:"id" db:"id"`
	Name             string    `json:"name" db:"name"`
	Description      string    `json:"description" db:"description"`
	EmbeddingBaseURL string    `json:"embedding_base_url" db:"embedding_base_url"`
	EmbeddingAPIKey  string    `json:"-" db:"embedding_api_key"`
	EmbeddingModel   string    `json:"embedding_model" db:"embedding_model"`
	ChunkSize        int       `json:"chunk_size" db:"chunk_size"`
	ChunkOverlap     int       `json:"chunk_overlap" db:"chunk_overlap"`
	Dimension        int       `json:"dimension" db:"dimension"`
	Status           string    `json:"status" db:"status"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

type Document struct {
	ID              int64     `json:"id" db:"id"`
	KnowledgeBaseID int64     `json:"knowledge_base_id" db:"knowledge_base_id"`
	Name            string    `json:"name" db:"name"`
	Content         string    `json:"-" db:"content"`
	CharCount       int       `json:"char_count" db:"char_count"`
	ChunkCount      int       `json:"chunk_count" db:"chunk_count"`
	Status          string    `json:"status" db:"status"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type Chunk struct {
	ID              int64     `json:"id" db:"id"`
	DocumentID      int64     `json:"document_id" db:"document_id"`
	KnowledgeBaseID int64     `json:"knowledge_base_id" db:"knowledge_base_id"`
	ChunkIndex      int       `json:"chunk_index" db:"chunk_index"`
	Content         string    `json:"content" db:"content"`
	Embedding       []byte    `json:"-" db:"embedding"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type AgentKnowledgeBase struct {
	AgentID         int64     `json:"agent_id" db:"agent_id"`
	KnowledgeBaseID int64     `json:"knowledge_base_id" db:"knowledge_base_id"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

const (
	KBStatusActive   = "active"
	KBStatusInactive = "inactive"

	DocStatusPending    = "pending"
	DocStatusProcessing = "processing"
	DocStatusCompleted  = "completed"
	DocStatusFailed     = "failed"

	ToolTypeKnowledgeSearch = "knowledge_search"
)
