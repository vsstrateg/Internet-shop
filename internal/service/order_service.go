package service

import "context"

// OrderStorage представляет интерфейс для работы с хранилищем заказов
type OrderStorage interface {
	InsertOrder(ctx context.Context, email string) (int64, error)
}

// OrderService предоставляет методы для работы с заказами
type OrderService struct {
	storage OrderStorage
}

// NewOrderService создает новый экземпляр OrderService
func NewOrderService(storage OrderStorage) *OrderService {
	return &OrderService{storage: storage}
}

// CreateOrder создает новый заказ для пользователя с указанным email и возвращает его ID
func (s *OrderService) CreateOrder(ctx context.Context, email string) (int64, error) {
	return s.storage.InsertOrder(ctx, email)
}
