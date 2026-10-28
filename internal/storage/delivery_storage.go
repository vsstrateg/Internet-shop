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
)

func NewDeliveryStorage(db *sql.DB) *DeliveryStorage {
	return &DeliveryStorage{db: db}
}

func (s *DeliveryStorage) SaveDelivery(ctx context.Context, orderID int64, carrier string, status string) error {
	query := `INSERT INTO deliveries (order_id, carrier, status) VALUES ($1, $2, $3) ON CONFLICT (order_id) DO NOTHING`
	res, err := s.db.ExecContext(ctx, query, orderID, carrier, status)
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

func (s *DeliveryStorage) GetDeliveryByOrderID(ctx context.Context, orderID int64) (string, error) {
	var status string

	query := `SELECT status FROM deliveries WHERE order_id = $1`
	err := s.db.QueryRowContext(ctx, query, orderID).Scan(&status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrDeliveryNotFound
		}
		return "", ErrDatabase
	}
	return status, nil
}
