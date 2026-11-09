package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

// OrderService представляет интерфейс для работы с заказами
type OrderService interface {
	CreateOrder(ctx context.Context, email string) (int64, error)
}

// OrderHandler предоставляет методы для обработки HTTP-запросов, связанных с заказами
type OrderHandler struct {
	service OrderService
}

// OrderResponse представляет структуру ответа на запрос создания заказа
type OrderResponse struct {
	Status  string `json:"status"`
	Email   string `json:"email"`
	OrderID int64  `json:"order_id"`
}

// NewOrderHandler создает новый экземпляр OrderHandler с указанным сервисом заказов
func NewOrderHandler(service OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

// CreateOrderHandler обрабатывает HTTP-запрос на создание нового заказа для авторизованного пользователя
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
	var response = OrderResponse{
		Status:  "success",
		Email:   email,
		OrderID: orderID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Warn("клиент оборвал соединение", "error", err)
		return
	}
}
