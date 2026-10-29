package storage

import (
	"context"
	"database/sql"
	"errors"
)

type DeliveryStorage struct {
	db *sql.DB
}

var (
	ErrDeliveryAlreadyExists = errors.New("Order already exists")
	ErrDeliveryNotFound      = errors.New("Order doesn't exists")
	ErrOrderNotFound         = errors.New("order not found")
	ErrOrderForbidden        = errors.New("order belongs to another user")
)

func NewDeliveryStorage(db *sql.DB) *DeliveryStorage {
	return &DeliveryStorage{db: db}
}

func (s *DeliveryStorage) SaveDelivery(ctx context.Context, email string, orderID int64, carrier string, status string) error {
	query1 := `SELECT user_email FROM orders WHERE id = $1`

	var orderEmail string
	err := s.db.QueryRowContext(ctx, query1, orderID).Scan(&orderEmail)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrOrderNotFound
		}
		return ErrDatabase
	}
	if orderEmail != email {
		return ErrOrderForbidden
	}

	query2 := `INSERT INTO deliveries (order_id, carrier, status) VALUES ($1, $2, $3) ON CONFLICT (order_id) DO NOTHING`

	res, err := s.db.ExecContext(ctx, query2, orderID, carrier, status)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrDeliveryAlreadyExists
	}
	return nil
}

func (s *DeliveryStorage) GetDeliveryByOrderID(ctx context.Context, email string, orderID int64) (string, error) {
	var status string
	var orderEmail string

	query := `SELECT orders.user_email, deliveries.status FROM deliveries JOIN orders ON deliveries.order_id = orders.id WHERE deliveries.order_id = $1`
	err := s.db.QueryRowContext(ctx, query, orderID).Scan(&orderEmail, &status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrDeliveryNotFound
		}
		return "", ErrDatabase
	}
	if orderEmail != email {
		return "", ErrOrderForbidden
	}
	return status, nil
}
