// Package service предоставляет сервисы для аутентификации пользователей и управления токенами
package service

import (
	"auth-service/internal/storage"
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// UserStorage представляет интерфейс для работы с хранилищем пользователей
type UserStorage interface {
	GetPasswordHash(ctx context.Context, email string) (string, error)
	SaveUser(ctx context.Context, email string, passwordHash []byte) error
}

// TokenStorage представляет интерфейс для работы с хранилищем токенов
type TokenStorage interface {
	SaveToken(ctx context.Context, email string, token string, ttl time.Duration) error
	GetToken(ctx context.Context, email string) (string, error)
	IncrementRequests(ctx context.Context, key string, ttl time.Duration) (int64, error)
}

// AuthService предоставляет методы для аутентификации пользователей, включая регистрацию, вход в систему и проверку токенов
type AuthService struct {
	storage   UserStorage
	token     TokenStorage
	jwtSecret []byte
}

// NewAuthService создает новый экземпляр AuthService с указанным хранилищем пользователей, хранилищем токенов и секретным ключом для JWT
func NewAuthService(storage UserStorage, token TokenStorage, jwtSecret []byte) *AuthService {
	return &AuthService{
		storage:   storage,
		token:     token,
		jwtSecret: jwtSecret,
	}
}

// Register создает нового пользователя, сохраняет его учетные данные и возвращает JWT токен
func (a *AuthService) Register(ctx context.Context, email string, password string) (string, error) {

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	if err := a.storage.SaveUser(ctx, email, bytes); err != nil {
		return "", err
	}

	token, err := a.GenerateJWT(email)
	if err != nil {
		return "", err
	}

	if err := a.token.SaveToken(ctx, email, token, 20*time.Minute); err != nil {
		return "", err
	}

	return token, nil
}

// Login проверяет учетные данные пользователя и возвращает JWT токен, если они верны
func (a *AuthService) Login(ctx context.Context, email string, password string) (string, error) {
	hash, err := a.storage.GetPasswordHash(ctx, email)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return "", storage.ErrInvalidCredentials
	}

	if token, err := a.token.GetToken(ctx, email); err == nil && token != "" {
		return token, nil
	}

	token, err := a.GenerateJWT(email)
	if err != nil {
		return "", err
	}

	if err := a.token.SaveToken(ctx, email, token, 20*time.Minute); err != nil {
		return "", err
	}

	return token, nil

}

// ValidateToken проверяет JWT токен и возвращает email пользователя, если токен действителен
func (a *AuthService) ValidateToken(ctx context.Context, tokenString string) (string, error) {
	email, err := a.ParseJWT(tokenString)
	if err != nil {
		return "", err
	}

	cachedToken, err := a.token.GetToken(ctx, email)
	if err != nil {
		return "", err
	}

	if cachedToken != tokenString {
		return "", storage.ErrTokenNotFound
	}

	return email, nil
}

// AllowUserRequest проверяет, не превысил ли пользователь лимит запросов
func (a *AuthService) AllowUserRequest(ctx context.Context, email string) (bool, error) {

	key := "rate:orders:" + email

	count, err := a.token.IncrementRequests(ctx, key, 1*time.Minute)
	if err != nil {
		return false, err
	}
	if count > 4 {
		return false, nil
	}

	return true, nil
}
