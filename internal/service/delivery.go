package service

import (
	"context"
)

type DeliveryService struct {
}

func (d *DeliveryService) SelectDelivery(ctx context.Context, orderID string, carrier string) (string, error) {
	return "", nil
}

func (d *DeliveryService) CheckStatusDelivery(ctx context.Context, orderID string) (string, error) {
	return "", nil
}
