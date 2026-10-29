package handler

import (
	"context"
	"encoding/json"
	"net/http"
)

type OrderService interface {
	CreateOrder(ctx context.Context, email string) (int64, error)
}

type OrderHandler struct {
	service OrderService
}

func NewOrderHandler(service OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	email, ok := r.Context().Value(UserEmailKey).(string)
	if !ok || email == "" {
		sendError(w, http.StatusUnauthorized, "Вы не авторизованы")
		return
	}

	orderID, err := h.service.CreateOrder(r.Context(), email)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Ошибка создания заказа")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"status":   "success",
		"email":    email,
		"order_id": orderID,
	})
}
