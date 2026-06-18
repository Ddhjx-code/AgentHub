package transaction

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Ddhjx-code/AgentHub/internal/model"
)

type Repository interface {
	Create(ctx context.Context, tx *model.Transaction) error
	ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]*model.Transaction, int, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, tx *model.Transaction) error {
	query := `INSERT INTO transactions (user_id, type, agent_id, agent_name, amount, status, note, created_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	now := time.Now()
	result, err := r.db.ExecContext(ctx, query,
		tx.UserID, tx.Type, tx.AgentID, tx.AgentName,
		tx.Amount, tx.Status, tx.Note, now)
	if err != nil {
		return fmt.Errorf("insert transaction: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}
	tx.ID = id
	tx.CreatedAt = now
	return nil
}

func (r *repository) ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]*model.Transaction, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM transactions WHERE user_id = ?`, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count transactions: %w", err)
	}

	query := `SELECT id, user_id, type, agent_id, agent_name, amount, status, note, created_at
	          FROM transactions WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list transactions: %w", err)
	}
	defer rows.Close()

	var txs []*model.Transaction
	for rows.Next() {
		tx := &model.Transaction{}
		err := rows.Scan(&tx.ID, &tx.UserID, &tx.Type, &tx.AgentID, &tx.AgentName,
			&tx.Amount, &tx.Status, &tx.Note, &tx.CreatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("scan transaction: %w", err)
		}
		txs = append(txs, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return txs, total, nil
}
