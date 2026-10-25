package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"auth-service/internal/storage"
)

type AuthService interface {
	Register(ctx context.Context, email string, password string) (string, error)
	Login(ctx context.Context, email string, password string) (string, error)
	ValidateToken(ctx context.Context, tokenString string) (string, error)
	AllowUserRequest(ctx context.Context, email string) (bool, error)
}

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthHandler struct {
	service AuthService
}

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

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var cred Credentials

	if err := json.NewDecoder(r.Body).Decode(&cred); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	if len(cred.Email) < 8 || len(cred.Password) < 8 {
		sendError(w, http.StatusBadRequest, "Длина пароля или почты меньше 8 символов")
		return
	}

	token, err := h.service.Register(r.Context(), cred.Email, cred.Password)
	if err != nil {
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			sendError(w, http.StatusConflict, "User already exists")
			return
		}
		sendError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	setTokenCookie(w, token)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"email": cred.Email,
		"token": token,
	})

	log.Printf("Регистрация Успешна: %s", cred.Email)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var cred Credentials

	if err := json.NewDecoder(r.Body).Decode(&cred); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	if len(cred.Email) < 8 || len(cred.Password) < 8 {
		sendError(w, http.StatusBadRequest, "Длина пароля или почты меньше 8 символов")
		return
	}

	token, err := h.service.Login(r.Context(), cred.Email, cred.Password)
	if err != nil {
		if errors.Is(err, storage.ErrDatabase) {
			sendError(w, http.StatusInternalServerError, "Internal server error")
		} else {
			sendError(w, http.StatusUnauthorized, err.Error())
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	setTokenCookie(w, token)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"email": cred.Email,
		"token": token,
	})
}
