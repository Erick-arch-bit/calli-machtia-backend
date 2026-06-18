package services

import (
	"context"
	"fmt"
	"time"

	"github.com/calli-machtia/backend/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	cfg       *config.Config
	redis     *redis.Client
}

func NewAuthService(cfg *config.Config, rdb *redis.Client) *AuthService {
	return &AuthService{cfg: cfg, redis: rdb}
}

func (s *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(bytes), nil
}

func (s *AuthService) CheckPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

type AccessClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func (s *AuthService) GenerateAccessToken(userID, role string) (string, error) {
	claims := AccessClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.JWT_ACCESS_EXPIRY)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.JWT_SECRET))
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return signed, nil
}

func (s *AuthService) GenerateRefreshToken(userID string) (string, error) {
	claims := RefreshClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.JWT_REFRESH_EXPIRY)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.JWT_SECRET))
	if err != nil {
		return "", fmt.Errorf("sign refresh token: %w", err)
	}

	if err := s.StoreRefreshToken(userID, signed); err != nil {
		return "", err
	}

	return signed, nil
}

func (s *AuthService) ValidateAccessToken(tokenStr string) (*AccessClaims, error) {
	claims := &AccessClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.cfg.JWT_SECRET), nil
	})
	if err != nil {
		return nil, fmt.Errorf("validate access token: %w", err)
	}
	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}

func (s *AuthService) ValidateRefreshToken(tokenStr string) (*RefreshClaims, error) {
	claims := &RefreshClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.cfg.JWT_SECRET), nil
	})
	if err != nil {
		return nil, fmt.Errorf("validate refresh token: %w", err)
	}
	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	valid, err := s.ValidateRefreshTokenInRedis(claims.UserID, tokenStr)
	if err != nil || !valid {
		return nil, fmt.Errorf("refresh token revoked or not found")
	}

	return claims, nil
}

func (s *AuthService) StoreRefreshToken(userID, token string) error {
	if s.redis == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := "refresh_token:" + userID
	duration := s.cfg.JWT_REFRESH_EXPIRY
	err := s.redis.Set(ctx, key, token, duration).Err()
	if err != nil {
		return fmt.Errorf("store refresh token: %w", err)
	}
	return nil
}

func (s *AuthService) ValidateRefreshTokenInRedis(userID, token string) (bool, error) {
	if s.redis == nil {
		return true, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := "refresh_token:" + userID
	stored, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, fmt.Errorf("get refresh token: %w", err)
	}
	return stored == token, nil
}

func (s *AuthService) InvalidateRefreshToken(userID string) error {
	if s.redis == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := "refresh_token:" + userID
	return s.redis.Del(ctx, key).Err()
}
