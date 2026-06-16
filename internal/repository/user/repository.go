package user

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Ddhjx-code/AgentHub/internal/model"
)

type Repository interface {
	Create(ctx context.Context, user *model.User) error
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (email, name, password, avatar, status, created_at, updated_at)
              VALUES (?, ?, ?, ?, ?, ?, ?)`
	now := time.Now()
	result, err := r.db.ExecContext(ctx, query,
		user.Email, user.Name, user.Password, user.Avatar, user.Status, now, now)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}
	user.ID = id
	user.CreatedAt = now
	user.UpdatedAt = now
	return nil
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, email, name, password, avatar, status, created_at, updated_at
              FROM users WHERE email = ?`
	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Name, &user.Password,
		&user.Avatar, &user.Status, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query user by email: %w", err)
	}
	return user, nil
}

func (r *repository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	query := `SELECT id, email, name, password, avatar, status, created_at, updated_at
              FROM users WHERE id = ?`
	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Name, &user.Password,
		&user.Avatar, &user.Status, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query user by id: %w", err)
	}
	return user, nil
}
