// Package handler содержит обработчики HTTP-запросов для аутентификации и доставки
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"auth-service/internal/storage"
)

// AuthService представляет интерфейс для работы с аутентификацией
type AuthService interface {
	Register(ctx context.Context, email string, password string) (string, error)
	Login(ctx context.Context, email string, password string) (string, error)
	ValidateToken(ctx context.Context, tokenString string) (string, error)
	AllowUserRequest(ctx context.Context, email string) (bool, error)
}

// Credentials представляет структуру данных для аутентификации пользователя
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse представляет структуру ответа на запрос аутентификации
type AuthResponse struct {
	Email string `json:"email"`
}

// AuthHandler предоставляет методы для обработки HTTP-запросов, связанных с аутентификацией
type AuthHandler struct {
	service AuthService
}

// NewAuthHandler создает новый экземпляр AuthHandler с указанным сервисом аутентификации
func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func setTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   1200,
	})
}

// RegisterHandler обрабатывает HTTP-запрос на регистрацию нового пользователя
func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var cred Credentials

	if err := json.NewDecoder(r.Body).Decode(&cred); err != nil {
		sendError(w, http.StatusBadRequest, "Неверный формат JSON")
		return
	}

	if len(cred.Email) < 8 || len(cred.Password) < 8 {
		sendError(w, http.StatusBadRequest, "Длина пароля или почты меньше 8 символов")
		return
	}

	token, err := h.service.Register(r.Context(), cred.Email, cred.Password)
	if err != nil {
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			sendError(w, http.StatusConflict, "Пользователь уже существует")
			return
		}
		sendError(w, http.StatusInternalServerError, "Внутренняя ошибка сервера")
		return
	}

	resp := AuthResponse{
		Email: cred.Email,
	}
	w.Header().Set("Content-Type", "application/json")
	setTokenCookie(w, token)
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Warn("клиент оборвал соединение", "error", err)
		return
	}

	slog.Info("Регистрация успешна", "email", cred.Email)
}

// LoginHandler обрабатывает HTTP-запрос на вход пользователя в систему
func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var cred Credentials

	if err := json.NewDecoder(r.Body).Decode(&cred); err != nil {
		sendError(w, http.StatusBadRequest, "Неверный формат JSON")
		return
	}

	if len(cred.Email) < 8 || len(cred.Password) < 8 {
		sendError(w, http.StatusBadRequest, "Длина пароля или почты меньше 8 символов")
		return
	}

	token, err := h.service.Login(r.Context(), cred.Email, cred.Password)
	if err != nil {
		if errors.Is(err, storage.ErrDatabase) {
			sendError(w, http.StatusInternalServerError, "Внутренняя ошибка сервера")
		} else {
			sendError(w, http.StatusUnauthorized, err.Error())
		}
		return
	}

	resp := AuthResponse{
		Email: cred.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	setTokenCookie(w, token)
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Warn("клиент оборвал соединение", "error", err)
		return
	}
}
