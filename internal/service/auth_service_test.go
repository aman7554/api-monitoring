package service

import (
	"testing"
	"time"

	"pulsewatch/internal/config"
	"pulsewatch/internal/domain"

	"github.com/google/uuid"
)

func TestPasswordHashing(t *testing.T) {
	password := "Secret123!"
	hash, err := domain.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if !domain.CheckPasswordHash(password, hash) {
		t.Fatalf("password hash check failed for valid password")
	}

	if domain.CheckPasswordHash("WrongPass", hash) {
		t.Fatalf("password hash check succeeded for invalid password")
	}
}

func TestJWTTokenGenerationAndValidation(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:        "test-secret-key-123456",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}

	authSvc := NewAuthService(nil, nil, cfg)
	user := &domain.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  domain.RoleUser,
	}

	tokens, err := authSvc.GenerateTokens(user)
	if err != nil {
		t.Fatalf("failed to generate tokens: %v", err)
	}

	claims, err := authSvc.ValidateToken(tokens.AccessToken)
	if err != nil {
		t.Fatalf("failed to validate valid access token: %v", err)
	}

	if claims.Email != user.Email {
		t.Fatalf("expected email %s, got %s", user.Email, claims.Email)
	}
}
