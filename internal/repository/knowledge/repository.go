package knowledge

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Ddhjx-code/AgentHub/internal/model"
)

type Repository interface {
	CreateKB(ctx context.Context, kb *model.KnowledgeBase) error
	GetKBByID(ctx context.Context, id int64) (*model.KnowledgeBase, error)
	ListKBs(ctx context.Context) ([]*model.KnowledgeBase, error)
	UpdateKB(ctx context.Context, kb *model.KnowledgeBase) error
	DeleteKB(ctx context.Context, id int64) error
	UpdateKBDimension(ctx context.Context, id int64, dimension int) error

	CreateDocument(ctx context.Context, doc *model.Document) error
	GetDocumentByID(ctx context.Context, id int64) (*model.Document, error)
	ListDocumentsByKBID(ctx context.Context, kbID int64) ([]*model.Document, error)
	DeleteDocument(ctx context.Context, id int64) error
	UpdateDocumentStatus(ctx context.Context, id int64, status string, chunkCount int) error

	BindAgentKB(ctx context.Context, agentID, kbID int64) error
	UnbindAgentKB(ctx context.Context, agentID, kbID int64) error
	ListKBsByAgentID(ctx context.Context, agentID int64) ([]*model.KnowledgeBase, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateKB(ctx context.Context, kb *model.KnowledgeBase) error {
	query := `INSERT INTO knowledge_bases (name, description, embedding_base_url, embedding_api_key,
	           embedding_model, chunk_size, chunk_overlap, dimension, status, created_at, updated_at)
	           VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	now := time.Now()
	result, err := r.db.ExecContext(ctx, query,
		kb.Name, kb.Description, kb.EmbeddingBaseURL, kb.EmbeddingAPIKey,
		kb.EmbeddingModel, kb.ChunkSize, kb.ChunkOverlap, kb.Dimension, kb.Status, now, now)
	if err != nil {
		return fmt.Errorf("insert knowledge base: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}
	kb.ID = id
	kb.CreatedAt = now
	kb.UpdatedAt = now
	return nil
}

func (r *repository) GetKBByID(ctx context.Context, id int64) (*model.KnowledgeBase, error) {
	query := `SELECT id, name, description, embedding_base_url, embedding_api_key,
	           embedding_model, chunk_size, chunk_overlap, dimension, status, created_at, updated_at
	           FROM knowledge_bases WHERE id = ?`
	kb := &model.KnowledgeBase{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&kb.ID, &kb.Name, &kb.Description, &kb.EmbeddingBaseURL, &kb.EmbeddingAPIKey,
		&kb.EmbeddingModel, &kb.ChunkSize, &kb.ChunkOverlap, &kb.Dimension,
		&kb.Status, &kb.CreatedAt, &kb.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query knowledge base: %w", err)
	}
	return kb, nil
}

func (r *repository) ListKBs(ctx context.Context) ([]*model.KnowledgeBase, error) {
	query := `SELECT id, name, description, embedding_base_url, embedding_api_key,
	           embedding_model, chunk_size, chunk_overlap, dimension, status, created_at, updated_at
	           FROM knowledge_bases ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list knowledge bases: %w", err)
	}
	defer rows.Close()

	var kbs []*model.KnowledgeBase
	for rows.Next() {
		kb := &model.KnowledgeBase{}
		if err := rows.Scan(
			&kb.ID, &kb.Name, &kb.Description, &kb.EmbeddingBaseURL, &kb.EmbeddingAPIKey,
			&kb.EmbeddingModel, &kb.ChunkSize, &kb.ChunkOverlap, &kb.Dimension,
			&kb.Status, &kb.CreatedAt, &kb.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan knowledge base: %w", err)
		}
		kbs = append(kbs, kb)
	}
	return kbs, rows.Err()
}

func (r *repository) UpdateKB(ctx context.Context, kb *model.KnowledgeBase) error {
	query := `UPDATE knowledge_bases SET name=?, description=?, embedding_base_url=?, embedding_api_key=?,
	           embedding_model=?, chunk_size=?, chunk_overlap=?, status=?, updated_at=? WHERE id=?`
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		kb.Name, kb.Description, kb.EmbeddingBaseURL, kb.EmbeddingAPIKey,
		kb.EmbeddingModel, kb.ChunkSize, kb.ChunkOverlap, kb.Status, now, kb.ID)
	if err != nil {
		return fmt.Errorf("update knowledge base: %w", err)
	}
	kb.UpdatedAt = now
	return nil
}

func (r *repository) DeleteKB(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM knowledge_bases WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete knowledge base: %w", err)
	}
	return nil
}

func (r *repository) UpdateKBDimension(ctx context.Context, id int64, dimension int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE knowledge_bases SET dimension = ?, updated_at = ? WHERE id = ? AND dimension = 0`,
		dimension, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update dimension: %w", err)
	}
	return nil
}

func (r *repository) CreateDocument(ctx context.Context, doc *model.Document) error {
	query := `INSERT INTO documents (knowledge_base_id, name, content, char_count, chunk_count, status, created_at)
	           VALUES (?, ?, ?, ?, ?, ?, ?)`
	now := time.Now()
	result, err := r.db.ExecContext(ctx, query,
		doc.KnowledgeBaseID, doc.Name, doc.Content, doc.CharCount, doc.ChunkCount, doc.Status, now)
	if err != nil {
		return fmt.Errorf("insert document: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}
	doc.ID = id
	doc.CreatedAt = now
	return nil
}

func (r *repository) GetDocumentByID(ctx context.Context, id int64) (*model.Document, error) {
	query := `SELECT id, knowledge_base_id, name, content, char_count, chunk_count, status, created_at
	           FROM documents WHERE id = ?`
	doc := &model.Document{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&doc.ID, &doc.KnowledgeBaseID, &doc.Name, &doc.Content,
		&doc.CharCount, &doc.ChunkCount, &doc.Status, &doc.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query document: %w", err)
	}
	return doc, nil
}

func (r *repository) ListDocumentsByKBID(ctx context.Context, kbID int64) ([]*model.Document, error) {
	query := `SELECT id, knowledge_base_id, name, '', char_count, chunk_count, status, created_at
	           FROM documents WHERE knowledge_base_id = ? ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, kbID)
	if err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}
	defer rows.Close()

	var docs []*model.Document
	for rows.Next() {
		doc := &model.Document{}
		if err := rows.Scan(
			&doc.ID, &doc.KnowledgeBaseID, &doc.Name, &doc.Content,
			&doc.CharCount, &doc.ChunkCount, &doc.Status, &doc.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan document: %w", err)
		}
		docs = append(docs, doc)
	}
	return docs, rows.Err()
}

func (r *repository) DeleteDocument(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM documents WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete document: %w", err)
	}
	return nil
}

func (r *repository) UpdateDocumentStatus(ctx context.Context, id int64, status string, chunkCount int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE documents SET status = ?, chunk_count = ? WHERE id = ?`,
		status, chunkCount, id)
	if err != nil {
		return fmt.Errorf("update document status: %w", err)
	}
	return nil
}

func (r *repository) BindAgentKB(ctx context.Context, agentID, kbID int64) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO agent_knowledge_bases (agent_id, knowledge_base_id, created_at) VALUES (?, ?, ?)`,
		agentID, kbID, time.Now())
	if err != nil {
		return fmt.Errorf("bind agent kb: %w", err)
	}
	return nil
}

func (r *repository) UnbindAgentKB(ctx context.Context, agentID, kbID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM agent_knowledge_bases WHERE agent_id = ? AND knowledge_base_id = ?`,
		agentID, kbID)
	if err != nil {
		return fmt.Errorf("unbind agent kb: %w", err)
	}
	return nil
}

func (r *repository) ListKBsByAgentID(ctx context.Context, agentID int64) ([]*model.KnowledgeBase, error) {
	query := `SELECT kb.id, kb.name, kb.description, kb.embedding_base_url, kb.embedding_api_key,
	           kb.embedding_model, kb.chunk_size, kb.chunk_overlap, kb.dimension, kb.status,
	           kb.created_at, kb.updated_at
	           FROM knowledge_bases kb
	           JOIN agent_knowledge_bases akb ON akb.knowledge_base_id = kb.id
	           WHERE akb.agent_id = ?
	           ORDER BY akb.created_at`
	rows, err := r.db.QueryContext(ctx, query, agentID)
	if err != nil {
		return nil, fmt.Errorf("list agent kbs: %w", err)
	}
	defer rows.Close()

	var kbs []*model.KnowledgeBase
	for rows.Next() {
		kb := &model.KnowledgeBase{}
		if err := rows.Scan(
			&kb.ID, &kb.Name, &kb.Description, &kb.EmbeddingBaseURL, &kb.EmbeddingAPIKey,
			&kb.EmbeddingModel, &kb.ChunkSize, &kb.ChunkOverlap, &kb.Dimension,
			&kb.Status, &kb.CreatedAt, &kb.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan agent kb: %w", err)
		}
		kbs = append(kbs, kb)
	}
	return kbs, rows.Err()
}
