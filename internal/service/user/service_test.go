package user

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/Ddhjx-code/AgentHub/internal/config"
	"github.com/Ddhjx-code/AgentHub/internal/database"
	"github.com/Ddhjx-code/AgentHub/internal/model"
	"github.com/Ddhjx-code/AgentHub/pkg/errcode"
	userRepo "github.com/Ddhjx-code/AgentHub/internal/repository/user"
	walletRepo "github.com/Ddhjx-code/AgentHub/internal/repository/wallet"
)

func setupTestService(t *testing.T) Service {
	t.Helper()
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ur := userRepo.NewRepository(db)
	wr := walletRepo.NewRepository(db)
	jwtCfg := config.JWTConfig{Secret: "test-secret", ExpireHour: 1}
	lg := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	return NewService(ur, wr, jwtCfg, lg)
}

func TestRegister(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	user, err := svc.Register(ctx, RegisterRequest{
		Email:    "reg@example.com",
		Name:     "RegUser",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if user.ID == 0 {
		t.Fatal("expected non-zero user ID")
	}
	if user.Email != "reg@example.com" {
		t.Fatalf("expected email reg@example.com, got %s", user.Email)
	}
	if user.Status != model.UserStatusActive {
		t.Fatalf("expected status active, got %s", user.Status)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	svc.Register(ctx, RegisterRequest{
		Email: "dup@example.com", Name: "First", Password: "password123",
	})

	_, err := svc.Register(ctx, RegisterRequest{
		Email: "dup@example.com", Name: "Second", Password: "password456",
	})
	if err == nil {
		t.Fatal("expected error for duplicate email")
	}
	if err != errcode.ErrEmailExists {
		t.Fatalf("expected ErrEmailExists, got %v", err)
	}
}

func TestLoginSuccess(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	svc.Register(ctx, RegisterRequest{
		Email: "login@example.com", Name: "Login", Password: "password123",
	})

	token, user, err := svc.Login(ctx, LoginRequest{
		Email: "login@example.com", Password: "password123",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
	if user.Email != "login@example.com" {
		t.Fatalf("expected email login@example.com, got %s", user.Email)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	svc.Register(ctx, RegisterRequest{
		Email: "wrong@example.com", Name: "Wrong", Password: "password123",
	})

	_, _, err := svc.Login(ctx, LoginRequest{
		Email: "wrong@example.com", Password: "wrongpassword",
	})
	if err != errcode.ErrInvalidPassword {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestLoginNonExistentUser(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	_, _, err := svc.Login(ctx, LoginRequest{
		Email: "ghost@example.com", Password: "password123",
	})
	if err != errcode.ErrInvalidPassword {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestGetProfile(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	created, _ := svc.Register(ctx, RegisterRequest{
		Email: "profile@example.com", Name: "Profile", Password: "password123",
	})

	user, err := svc.GetProfile(ctx, created.ID)
	if err != nil {
		t.Fatalf("get profile: %v", err)
	}
	if user.Email != "profile@example.com" {
		t.Fatalf("expected email profile@example.com, got %s", user.Email)
	}
}

func TestGetProfileNotFound(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	_, err := svc.GetProfile(ctx, 9999)
	if err != errcode.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
