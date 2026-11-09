package service

import (
	"context"
)

// DeliveryStorage представляет интерфейс для работы с хранилищем информации о доставке
type DeliveryStorage interface {
	SaveDelivery(ctx context.Context, email string, orderID int64, carrier string, status string) error
	GetDeliveryByOrderID(ctx context.Context, email string, orderID int64) (string, error)
}

// DeliveryService предоставляет методы для работы с информацией о доставке заказов
type DeliveryService struct {
	storage DeliveryStorage
}

// NewDeliveryService создает новый экземпляр DeliveryService
func NewDeliveryService(storage DeliveryStorage) *DeliveryService {
	return &DeliveryService{
		storage: storage,
	}
}

// SelectDelivery выбирает доставку для заказа и сохраняет информацию о ней в хранилище
func (d *DeliveryService) SelectDelivery(ctx context.Context, email string, orderID int64, carrier string) (string, error) {

	status := "created"
	if err := d.storage.SaveDelivery(ctx, email, orderID, carrier, status); err != nil {
		return "", err
	}

	return status, nil
}

// CheckStatusDelivery проверяет статус доставки для заказа и возвращает его
func (d *DeliveryService) CheckStatusDelivery(ctx context.Context, email string, orderID int64) (string, error) {

	status, err := d.storage.GetDeliveryByOrderID(ctx, email, orderID)
	if err != nil {
		return "", err
	}
	return status, nil
}
