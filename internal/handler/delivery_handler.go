package handler

import (
	"auth-service/internal/storage"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
)

// DeliveryService представляет интерфейс для работы с доставкой
type DeliveryService interface {
	SelectDelivery(ctx context.Context, email string, orderID int64, carrier string) (string, error)
	CheckStatusDelivery(ctx context.Context, email string, orderID int64) (string, error)
}

// DeliveryCred представляет структуру данных для выбора доставки
type DeliveryCred struct {
	OrderID int64  `json:"order_id,string"`
	Carrier string `json:"carrier"`
}

// SelectDeliveryResponse представляет структуру ответа на запрос выбора доставки
type SelectDeliveryResponse struct {
	DeliveryCred
	Status string `json:"status"`
}

// DeliveryStatusResponse представляет структуру ответа на запрос проверки статуса доставки
type DeliveryStatusResponse struct {
	OrderID int64  `json:"order_id"`
	Status  string `json:"status"`
}

// DeliveryHandler предоставляет методы для обработки HTTP-запросов, связанных с доставкой
type DeliveryHandler struct {
	delivery DeliveryService
}

// NewDeliveryHandler создает новый экземпляр DeliveryHandler с указанным сервисом доставки
func NewDeliveryHandler(delivery DeliveryService) *DeliveryHandler {
	return &DeliveryHandler{delivery: delivery}
}

// SelectDeliveryHandler обрабатывает HTTP-запрос на выбор доставки для заказа авторизованного пользователя
func (d *DeliveryHandler) SelectDeliveryHandler(w http.ResponseWriter, r *http.Request) {
	email, ok := r.Context().Value(UserEmailKey).(string)
	if !ok || email == "" {
		sendError(w, http.StatusUnauthorized, "Вы не авторизованы")
		return
	}

	var delcred DeliveryCred

	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "Разрешен только метод POST")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	defer func() {
		if err := r.Body.Close(); err != nil {
			slog.Error("ошибка закрытия тела запроса", "error", err)
		}
	}()

	if err := json.NewDecoder(r.Body).Decode(&delcred); err != nil {
		sendError(w, http.StatusBadRequest, "Неверный формат JSON")
		return
	}

	if delcred.OrderID <= 0 || (delcred.Carrier != "cdek" && delcred.Carrier != "yandex_delivery" && delcred.Carrier != "boxberry") {
		sendError(w, http.StatusBadRequest, "Неверный ID заказа или Транспортная компания")
		return
	}

	status, err := d.delivery.SelectDelivery(r.Context(), email, delcred.OrderID, delcred.Carrier)
	if err != nil {
		if errors.Is(err, storage.ErrDeliveryAlreadyExists) {
			sendError(w, http.StatusConflict, "Доставка уже существует по этому заказу")
			return
		}
		if errors.Is(err, storage.ErrOrderNotFound) {
			sendError(w, http.StatusNotFound, "Заказ не найден")
			return
		}
		if errors.Is(err, storage.ErrOrderForbidden) {
			sendError(w, http.StatusForbidden, "Пожалуйста, выберите свой заказ")
			return
		}
		sendError(w, http.StatusInternalServerError, "Ошибка оформления доставки")
		return
	}
	resp := SelectDeliveryResponse{
		DeliveryCred: delcred,
		Status:       status,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Warn("клиент оборвал соединение", "error", err)
	}
}

// CheckStatusDeliveryHandler обрабатывает HTTP-запрос на проверку статуса доставки для заказа авторизованного пользователя
func (d *DeliveryHandler) CheckStatusDeliveryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "Разрешен только метод GET")
		return
	}

	email, ok := r.Context().Value(UserEmailKey).(string)
	if !ok || email == "" {
		sendError(w, http.StatusUnauthorized, "Вы не авторизованы")
		return
	}

	orderID, err := strconv.ParseInt(r.URL.Query().Get("order_id"), 10, 64)

	if err != nil || orderID <= 0 {
		sendError(w, http.StatusBadRequest, "Введите номер заказа")
		return
	}

	status, err := d.delivery.CheckStatusDelivery(r.Context(), email, orderID)
	if err != nil {
		if errors.Is(err, storage.ErrDeliveryNotFound) {
			sendError(w, http.StatusNotFound, "Доставка по заказу не найдена")
			return
		}
		if errors.Is(err, storage.ErrOrderForbidden) {
			sendError(w, http.StatusForbidden, "Пожалуйста, выберите свой заказ")
			return
		}
		sendError(w, http.StatusInternalServerError, "Ошибка проверки статуса доставки")
		return
	}

	resp := DeliveryStatusResponse{
		OrderID: orderID,
		Status:  status,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Warn("клиент оборвал соединение", "error", err)
	}
}
