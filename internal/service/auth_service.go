package service

import (
	"auth-service/internal/storage"
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserStorage interface {
	GetPasswordHash(email string) (string, error)
	SaveUser(email string, passwordHash []byte) error
}

type TokenStorage interface {
	SaveToken(ctx context.Context, email string, token string, ttl time.Duration) error
	GetToken(ctx context.Context, email string) (string, error)
	IncrementRequests(ctx context.Context, key string, ttl time.Duration) (int64, error)
}

type AuthService struct {
	storage   UserStorage
	token     TokenStorage
	jwtSecret []byte
}

func NewAuthService(storage UserStorage, token TokenStorage, jwtSecret []byte) *AuthService {
	return &AuthService{
		storage:   storage,
		token:     token,
		jwtSecret: jwtSecret,
	}
}

func (a *AuthService) Register(ctx context.Context, email string, password string) (string, error) {

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	if err := a.storage.SaveUser(email, bytes); err != nil {
		return "", err
	}

	token, err := a.GenerateJWT(email)
	if err != nil {
		return "", err
	}

	a.token.SaveToken(ctx, email, token, 20*time.Minute)

	return token, nil
}

func (a *AuthService) Login(ctx context.Context, email string, password string) (string, error) {
	hash, err := a.storage.GetPasswordHash(email)
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

// AllowUserRequest проверяет, не превысил ли пользователь лимит запросов (например, 100 в минуту).
// Возвращает true, если запрос разрешён, и false, если лимит превышен.
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
