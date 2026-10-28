package handler

import (
	"context"
	"encoding/json"
	"net/http"
)

type DeliveryService interface {
	SelectDelivery(ctx context.Context, orderID string, carrier string) (string, error)
	CheckStatusDelivery(ctx context.Context, orderID string) (string, error)
}

type DeliveryCred struct {
	OrderID string `json:"order_id"`
	Carrier string `json:"carrier"`
}

type DeliveryHandler struct {
	delivery DeliveryService
}

func NewDeliveryHandler(delivery DeliveryService) *DeliveryHandler {
	return &DeliveryHandler{delivery: delivery}
}

func (d *DeliveryHandler) SelectDeliveryHandler(w http.ResponseWriter, r *http.Request) {
	var delcred DeliveryCred

	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "Разрешен только метод POST")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	if err := json.NewDecoder(r.Body).Decode(&delcred); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}
	defer r.Body.Close()

	if delcred.OrderID == "" || (delcred.Carrier != "cdek" && delcred.Carrier != "yandex_delivery" && delcred.Carrier != "boxberry") {
		sendError(w, http.StatusBadRequest, "Неверный ID заказа или Транспортная компания")
		return
	}

	status, err := d.delivery.SelectDelivery(r.Context(), delcred.OrderID, delcred.Carrier)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Ошибка оформления доставки")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"order_id": delcred.OrderID,
		"carrier":  delcred.Carrier,
		"status":   status,
	})
}

func (d *DeliveryHandler) CheckStatusDeliveryHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "Разрешен только метод GET")
		return
	}

	orderID := r.URL.Query().Get("order_id")

	if orderID == "" {
		sendError(w, http.StatusBadRequest, "Введите номер заказа")
		return
	}

	status, err := d.delivery.CheckStatusDelivery(r.Context(), orderID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Ошибка проверки статуса доставки")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"order_id": orderID,
		"status":   status,
	})
}
