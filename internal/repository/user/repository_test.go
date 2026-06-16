package user

import (
	"context"
	"testing"

	"github.com/Ddhjx-code/AgentHub/internal/database"
	"github.com/Ddhjx-code/AgentHub/internal/model"
)

func setupTestDB(t *testing.T) Repository {
	t.Helper()
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewRepository(db)
}

func TestCreate(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	user := &model.User{
		Email:    "test@example.com",
		Name:     "Test",
		Password: "hashed",
		Status:   model.UserStatusActive,
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if user.ID == 0 {
		t.Fatal("expected non-zero ID after create")
	}
	if user.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be set")
	}
}

func TestGetByEmail(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	repo.Create(ctx, &model.User{
		Email: "find@example.com", Name: "Find", Password: "hashed", Status: model.UserStatusActive,
	})

	found, err := repo.GetByEmail(ctx, "find@example.com")
	if err != nil {
		t.Fatalf("get by email: %v", err)
	}
	if found == nil {
		t.Fatal("expected user, got nil")
	}
	if found.Name != "Find" {
		t.Fatalf("expected name Find, got %s", found.Name)
	}

	notFound, err := repo.GetByEmail(ctx, "nope@example.com")
	if err != nil {
		t.Fatalf("get by email (not found): %v", err)
	}
	if notFound != nil {
		t.Fatal("expected nil for non-existent email")
	}
}

func TestGetByID(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	user := &model.User{
		Email: "id@example.com", Name: "ID", Password: "hashed", Status: model.UserStatusActive,
	}
	repo.Create(ctx, user)

	found, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if found == nil || found.Email != "id@example.com" {
		t.Fatal("expected user with matching email")
	}

	notFound, err := repo.GetByID(ctx, 9999)
	if err != nil {
		t.Fatalf("get by id (not found): %v", err)
	}
	if notFound != nil {
		t.Fatal("expected nil for non-existent id")
	}
}

func TestDuplicateEmail(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	repo.Create(ctx, &model.User{
		Email: "dup@example.com", Name: "First", Password: "hashed", Status: model.UserStatusActive,
	})

	err := repo.Create(ctx, &model.User{
		Email: "dup@example.com", Name: "Second", Password: "hashed", Status: model.UserStatusActive,
	})
	if err == nil {
		t.Fatal("expected error for duplicate email")
	}
}
