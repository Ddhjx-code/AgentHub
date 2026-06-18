package conversation

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Ddhjx-code/AgentHub/internal/model"
)

type Repository interface {
	Create(ctx context.Context, conv *model.Conversation) error
	GetByID(ctx context.Context, id int64) (*model.Conversation, error)
	ListByUserID(ctx context.Context, userID int64) ([]*model.Conversation, error)
	UpdateTitle(ctx context.Context, id int64, title string) error
	Delete(ctx context.Context, id int64) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, conv *model.Conversation) error {
	query := `INSERT INTO conversations (user_id, agent_id, title, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?)`
	now := time.Now()
	result, err := r.db.ExecContext(ctx, query,
		conv.UserID, conv.AgentID, conv.Title, now, now)
	if err != nil {
		return fmt.Errorf("insert conversation: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}
	conv.ID = id
	conv.CreatedAt = now
	conv.UpdatedAt = now
	return nil
}

func (r *repository) GetByID(ctx context.Context, id int64) (*model.Conversation, error) {
	query := `SELECT id, user_id, agent_id, title, created_at, updated_at
	          FROM conversations WHERE id = ?`
	conv := &model.Conversation{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&conv.ID, &conv.UserID, &conv.AgentID, &conv.Title,
		&conv.CreatedAt, &conv.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query conversation: %w", err)
	}
	return conv, nil
}

func (r *repository) ListByUserID(ctx context.Context, userID int64) ([]*model.Conversation, error) {
	query := `SELECT id, user_id, agent_id, title, created_at, updated_at
	          FROM conversations WHERE user_id = ? ORDER BY updated_at DESC`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}
	defer rows.Close()

	var convs []*model.Conversation
	for rows.Next() {
		conv := &model.Conversation{}
		err := rows.Scan(&conv.ID, &conv.UserID, &conv.AgentID, &conv.Title,
			&conv.CreatedAt, &conv.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan conversation: %w", err)
		}
		convs = append(convs, conv)
	}
	return convs, rows.Err()
}

func (r *repository) UpdateTitle(ctx context.Context, id int64, title string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE conversations SET title = ?, updated_at = ? WHERE id = ?`,
		title, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update conversation title: %w", err)
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM conversations WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete conversation: %w", err)
	}
	return nil
}
