package service

import (
	"context"
)

type DeliveryStorage interface {
	SaveDelivery(ctx context.Context, email string, orderID int64, carrier string, status string) error
	GetDeliveryByOrderID(ctx context.Context, email string, orderID int64) (string, error)
}

type DeliveryService struct {
	storage DeliveryStorage
}

func NewDeliveryService(storage DeliveryStorage) *DeliveryService {
	return &DeliveryService{
		storage: storage,
	}
}

func (d *DeliveryService) SelectDelivery(ctx context.Context, email string, orderID int64, carrier string) (string, error) {

	status := "created"
	if err := d.storage.SaveDelivery(ctx, email, orderID, carrier, status); err != nil {
		return "", err
	}

	return status, nil
}

func (d *DeliveryService) CheckStatusDelivery(ctx context.Context, email string, orderID int64) (string, error) {

	status, err := d.storage.GetDeliveryByOrderID(ctx, email, orderID)
	if err != nil {
		return "", err
	}
	return status, nil
}
