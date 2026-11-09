package storage

import (
	"context"
	"database/sql"
)

// OrderStorage представляет хранилище для информации о заказах
type OrderStorage struct {
	db *sql.DB
}

// NewOrderStorage создает новый экземпляр OrderStorage
func NewOrderStorage(db *sql.DB) *OrderStorage {
	return &OrderStorage{db: db}
}

// InsertOrder вставляет новый заказ в базу данных и возвращает его ID
func (s *OrderStorage) InsertOrder(ctx context.Context, email string) (int64, error) {
	var orderID int64

	query := `INSERT INTO orders (user_email, status) VALUES ($1, 'created') RETURNING id`

	err := s.db.QueryRowContext(ctx, query, email).Scan(&orderID)
	if err != nil {
		return 0, err
	}
	return orderID, nil
}
