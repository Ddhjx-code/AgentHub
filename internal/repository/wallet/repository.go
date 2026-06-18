package wallet

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Ddhjx-code/AgentHub/internal/model"
)

type Repository interface {
	Create(ctx context.Context, wallet *model.Wallet) error
	GetByUserID(ctx context.Context, userID int64) (*model.Wallet, error)
	Deduct(ctx context.Context, userID int64, amount int) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, wallet *model.Wallet) error {
	query := `INSERT INTO wallets (user_id, balance, updated_at) VALUES (?, ?, ?)`
	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, wallet.UserID, wallet.Balance, now)
	if err != nil {
		return fmt.Errorf("insert wallet: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}
	wallet.ID = id
	wallet.UpdatedAt = now
	return nil
}

func (r *repository) GetByUserID(ctx context.Context, userID int64) (*model.Wallet, error) {
	query := `SELECT id, user_id, balance, updated_at FROM wallets WHERE user_id = ?`
	w := &model.Wallet{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&w.ID, &w.UserID, &w.Balance, &w.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query wallet by user_id: %w", err)
	}
	return w, nil
}

func (r *repository) Deduct(ctx context.Context, userID int64, amount int) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE wallets SET balance = balance - ?, updated_at = ? WHERE user_id = ? AND balance >= ?`,
		amount, time.Now(), userID, amount)
	if err != nil {
		return fmt.Errorf("deduct wallet: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("insufficient balance")
	}
	return nil
}
