package handler

import (
	"context"
	"net/http"
)

type contextKey string

const UserEmailKey contextKey = "user_email"

func (h *AuthHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			sendError(w, http.StatusUnauthorized, "Не авторизован")
			return
		}

		email, err := h.service.ValidateToken(r.Context(), cookie.Value)
		if err != nil {
			sendError(w, http.StatusUnauthorized, "Невалидная сессия")
			return
		}

		ctx := context.WithValue(r.Context(), UserEmailKey, email)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *AuthHandler) UserRateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		email, ok := r.Context().Value(UserEmailKey).(string)
		if !ok || email == "" {
			sendError(w, http.StatusUnauthorized, "Вы не авторизованы")
			return
		}

		allowed, err := h.service.AllowUserRequest(r.Context(), email)
		if err != nil {
			sendError(w, http.StatusInternalServerError, "Сервер недоступен")
			return
		}
		if !allowed {
			sendError(w, http.StatusTooManyRequests, "Слишком много запросов, подождите")
			return
		}
		next.ServeHTTP(w, r)
	})
}
