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

func TestDeduct(t *testing.T) {
	repo, user := setupTestDB(t)
	ctx := context.Background()

	repo.Create(ctx, &model.Wallet{UserID: user.ID, Balance: 100})

	if err := repo.Deduct(ctx, user.ID, 30); err != nil {
		t.Fatalf("deduct: %v", err)
	}

	w, _ := repo.GetByUserID(ctx, user.ID)
	if w.Balance != 70 {
		t.Fatalf("expected balance 70, got %d", w.Balance)
	}
}

func TestDeductInsufficientBalance(t *testing.T) {
	repo, user := setupTestDB(t)
	ctx := context.Background()

	repo.Create(ctx, &model.Wallet{UserID: user.ID, Balance: 10})

	err := repo.Deduct(ctx, user.ID, 20)
	if err == nil {
		t.Fatal("expected error for insufficient balance")
	}

	w, _ := repo.GetByUserID(ctx, user.ID)
	if w.Balance != 10 {
		t.Fatalf("balance should be unchanged, got %d", w.Balance)
	}
}

func TestDeductExactBalance(t *testing.T) {
	repo, user := setupTestDB(t)
	ctx := context.Background()

	repo.Create(ctx, &model.Wallet{UserID: user.ID, Balance: 50})

	if err := repo.Deduct(ctx, user.ID, 50); err != nil {
		t.Fatalf("deduct exact balance: %v", err)
	}

	w, _ := repo.GetByUserID(ctx, user.ID)
	if w.Balance != 0 {
		t.Fatalf("expected balance 0, got %d", w.Balance)
	}
}
