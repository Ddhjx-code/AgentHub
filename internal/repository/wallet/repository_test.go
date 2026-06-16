package wallet

import (
	"context"
	"testing"

	"github.com/Ddhjx-code/AgentHub/internal/database"
	"github.com/Ddhjx-code/AgentHub/internal/model"
)

func setupTestDB(t *testing.T) (Repository, *model.User) {
	t.Helper()
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// Insert a user to satisfy foreign key
	result, err := db.Exec(
		`INSERT INTO users (email, name, password, status) VALUES (?, ?, ?, ?)`,
		"wallet@example.com", "Wallet", "hashed", "active",
	)
	if err != nil {
		t.Fatalf("insert test user: %v", err)
	}
	id, _ := result.LastInsertId()
	user := &model.User{ID: id}

	return NewRepository(db), user
}

func TestCreateAndGetByUserID(t *testing.T) {
	repo, user := setupTestDB(t)
	ctx := context.Background()

	wallet := &model.Wallet{
		UserID:  user.ID,
		Balance: 200,
	}

	if err := repo.Create(ctx, wallet); err != nil {
		t.Fatalf("create wallet: %v", err)
	}
	if wallet.ID == 0 {
		t.Fatal("expected non-zero wallet ID")
	}

	found, err := repo.GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("get by user_id: %v", err)
	}
	if found == nil {
		t.Fatal("expected wallet, got nil")
	}
	if found.Balance != 200 {
		t.Fatalf("expected balance 200, got %d", found.Balance)
	}
}

func TestGetByUserIDNotFound(t *testing.T) {
	repo, _ := setupTestDB(t)
	ctx := context.Background()

	found, err := repo.GetByUserID(ctx, 9999)
	if err != nil {
		t.Fatalf("get by user_id: %v", err)
	}
	if found != nil {
		t.Fatal("expected nil for non-existent user_id")
	}
}
