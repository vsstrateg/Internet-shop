// Package storage предоставляет хранилища для информации о доставке заказов
package storage

import (
	"context"
	"database/sql"
	"errors"
)

// DeliveryStorage представляет хранилище для информации о доставке заказов
type DeliveryStorage struct {
	db *sql.DB
}

var (
	// ErrDeliveryAlreadyExists указывает, что доставка для данного заказа уже существует
	ErrDeliveryAlreadyExists = errors.New("доставка для данного заказа уже существует")
	// ErrDeliveryNotFound указывает, что доставка для данного заказа не найдена
	ErrDeliveryNotFound = errors.New("доставка для данного заказа не найдена")
	// ErrOrderNotFound указывает, что заказ не найден
	ErrOrderNotFound = errors.New("заказ не найден")
	// ErrOrderForbidden указывает, что заказ принадлежит другому пользователю
	ErrOrderForbidden = errors.New("заказ принадлежит другому пользователю")
)

// NewDeliveryStorage создает новый экземпляр DeliveryStorage
func NewDeliveryStorage(db *sql.DB) *DeliveryStorage {
	return &DeliveryStorage{db: db}
}

// SaveDelivery сохраняет информацию о доставке для заказа в базе данных
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

// GetDeliveryByOrderID извлекает информацию о доставке для заказа по его ID и проверяет, принадлежит ли заказ пользователю с указанным email
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
