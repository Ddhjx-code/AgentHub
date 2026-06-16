package user

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Ddhjx-code/AgentHub/internal/config"
	"github.com/Ddhjx-code/AgentHub/internal/model"
	userRepo "github.com/Ddhjx-code/AgentHub/internal/repository/user"
	walletRepo "github.com/Ddhjx-code/AgentHub/internal/repository/wallet"
	"github.com/Ddhjx-code/AgentHub/pkg/errcode"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const registrationBonus = 200

type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*model.User, error)
	Login(ctx context.Context, req LoginRequest) (string, *model.User, error)
	GetProfile(ctx context.Context, userID int64) (*model.User, error)
}

type RegisterRequest struct {
	Email    string
	Name     string
	Password string
}

type LoginRequest struct {
	Email    string
	Password string
}

type service struct {
	userRepo   userRepo.Repository
	walletRepo walletRepo.Repository
	jwtCfg     config.JWTConfig
	logger     *slog.Logger
}

func NewService(
	ur userRepo.Repository,
	wr walletRepo.Repository,
	jwtCfg config.JWTConfig,
	logger *slog.Logger,
) Service {
	return &service{
		userRepo:   ur,
		walletRepo: wr,
		jwtCfg:     jwtCfg,
		logger:     logger,
	}
}

func (s *service) Register(ctx context.Context, req RegisterRequest) (*model.User, error) {
	existing, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("check email: %w", err)
	}
	if existing != nil {
		return nil, errcode.ErrEmailExists
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{
		Email:    req.Email,
		Name:     req.Name,
		Password: string(hashed),
		Avatar:   req.Name[:1],
		Status:   model.UserStatusActive,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	wallet := &model.Wallet{
		UserID:  user.ID,
		Balance: registrationBonus,
	}
	if err := s.walletRepo.Create(ctx, wallet); err != nil {
		s.logger.Error("failed to create wallet", "user_id", user.ID, "error", err)
	}

	s.logger.Info("user registered", "user_id", user.ID, "email", user.Email)
	return user, nil
}

func (s *service) Login(ctx context.Context, req LoginRequest) (string, *model.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return "", nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return "", nil, errcode.ErrInvalidPassword
	}
	if user.Status == model.UserStatusBanned {
		return "", nil, errcode.ErrUserBanned
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", nil, errcode.ErrInvalidPassword
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return "", nil, fmt.Errorf("generate token: %w", err)
	}

	return token, user, nil
}

func (s *service) GetProfile(ctx context.Context, userID int64) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, errcode.ErrNotFound
	}
	return user, nil
}

func (s *service) generateToken(userID int64) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   fmt.Sprintf("%d", userID),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(s.jwtCfg.ExpireHour) * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(now),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtCfg.Secret))
}
