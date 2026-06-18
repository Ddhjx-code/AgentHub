package message

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Ddhjx-code/AgentHub/internal/model"
)

type Repository interface {
	Create(ctx context.Context, msg *model.Message) error
	ListByConversationID(ctx context.Context, conversationID int64, limit int) ([]*model.Message, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, msg *model.Message) error {
	query := `INSERT INTO messages (conversation_id, role, content, tool_calls, tool_call_id, created_at)
	          VALUES (?, ?, ?, ?, ?, ?)`
	now := time.Now()
	result, err := r.db.ExecContext(ctx, query,
		msg.ConversationID, msg.Role, msg.Content,
		msg.ToolCalls, msg.ToolCallID, now)
	if err != nil {
		return fmt.Errorf("insert message: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}
	msg.ID = id
	msg.CreatedAt = now
	return nil
}

func (r *repository) ListByConversationID(ctx context.Context, conversationID int64, limit int) ([]*model.Message, error) {
	query := `SELECT id, conversation_id, role, content, tool_calls, tool_call_id, created_at
	          FROM messages WHERE conversation_id = ? ORDER BY id DESC LIMIT ?`
	rows, err := r.db.QueryContext(ctx, query, conversationID, limit)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	var msgs []*model.Message
	for rows.Next() {
		msg := &model.Message{}
		err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content,
			&msg.ToolCalls, &msg.ToolCallID, &msg.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		msgs = append(msgs, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}
