package handler

import (
	"encoding/json"
	"net/http"
)

func OrdersHandler(w http.ResponseWriter, r *http.Request) {
	email, ok := r.Context().Value(UserEmailKey).(string)
	if !ok {
		sendError(w, http.StatusUnauthorized, "Вы не авторизованы")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "success",
		"email":    email,
		"order_id": "258",
	})

}

func deliveryHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("delivery route"))
}
