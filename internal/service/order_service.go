package service

import "context"

type OrderStorage interface {
	InsertOrder(ctx context.Context, email string) (int64, error)
}

type OrderService struct {
	storage OrderStorage
}

func NewOrderService(storage OrderStorage) *OrderService {
	return &OrderService{storage: storage}
}

func (s *OrderService) CreateOrder(ctx context.Context, email string) (int64, error) {
	return s.storage.InsertOrder(ctx, email)
}
