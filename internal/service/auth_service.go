package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"pulsewatch/internal/config"
	"pulsewatch/internal/domain"
	"pulsewatch/internal/repository/postgres"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthService struct {
	userRepo *postgres.UserRepository
	keyRepo  *postgres.ApiKeyRepository
	cfg      *config.Config
}

func NewAuthService(userRepo *postgres.UserRepository, keyRepo *postgres.ApiKeyRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		keyRepo:  keyRepo,
		cfg:      cfg,
	}
}

type JWTClaims struct {
	UserID uuid.UUID       `json:"user_id"`
	Email  string          `json:"email"`
	Role   domain.UserRole `json:"role"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func (s *AuthService) Register(ctx context.Context, email, password, fullName string) (*domain.User, *TokenPair, error) {
	existing, _ := s.userRepo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, nil, domain.ErrEmailExists
	}

	hash, err := domain.HashPassword(password)
	if err != nil {
		return nil, nil, fmt.Errorf("password hashing failed: %w", err)
	}

	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hash,
		FullName:     fullName,
		Role:         domain.RoleUser,
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, nil, err
	}

	tokens, err := s.GenerateTokens(user)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*domain.User, *TokenPair, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, nil, domain.ErrInvalidCredentials
	}

	if !domain.CheckPasswordHash(password, user.PasswordHash) {
		return nil, nil, domain.ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, nil, domain.ErrUnauthorized
	}

	tokens, err := s.GenerateTokens(user)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (s *AuthService) GenerateTokens(user *domain.User) (*TokenPair, error) {
	now := time.Now()
	accessClaims := &JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.JWTAccessExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.ID.String(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, err
	}

	refreshClaims := &JWTClaims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.JWTRefreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.ID.String(),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
		ExpiresIn:    int64(s.cfg.JWTAccessExpiry.Seconds()),
	}, nil
}

func (s *AuthService) ValidateToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, domain.ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, domain.ErrInvalidToken
	}

	return claims, nil
}

func HashApiKey(rawKey string) string {
	h := sha256.New()
	h.Write([]byte(rawKey))
	return hex.EncodeToString(h.Sum(nil))
}

func (s *AuthService) CreateApiKey(ctx context.Context, projectID uuid.UUID, name string) (string, *domain.ApiKey, error) {
	rawToken := fmt.Sprintf("pw_%s_%s", uuid.New().String()[:8], uuid.New().String())
	prefix := rawToken[:8]
	hash := HashApiKey(rawToken)

	key := &domain.ApiKey{
		ID:        uuid.New(),
		ProjectID: projectID,
		Name:      name,
		KeyPrefix: prefix,
		KeyHash:   hash,
	}

	if err := s.keyRepo.Create(ctx, key); err != nil {
		return "", nil, err
	}

	return rawToken, key, nil
}

func (s *AuthService) ValidateApiKey(ctx context.Context, rawKey string) (*domain.ApiKey, error) {
	hash := HashApiKey(rawKey)
	key, err := s.keyRepo.GetByKeyHash(ctx, hash)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	if key.ExpiresAt != nil && time.Now().After(*key.ExpiresAt) {
		return nil, domain.ErrUnauthorized
	}

	_ = s.keyRepo.UpdateLastUsed(ctx, key.ID)
	return key, nil
}
