package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type UserStorage struct {
	db *sql.DB
}

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("неверный логин или пароль")
	ErrDatabase           = errors.New("ошибка базы данных")
)

func NewUserStorage(db *sql.DB) *UserStorage {
	return &UserStorage{db: db}
}

func NewPostgresDB(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open pgx: %w", err)
	}

	db.SetMaxOpenConns(25)                 // максимум открытых соединений
	db.SetMaxIdleConns(25)                 // соединений в простое
	db.SetConnMaxLifetime(5 * time.Minute) // ротация соединений

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return db, nil
}

func (s *UserStorage) SaveUser(email string, passwordHash []byte) error {
	query := `INSERT INTO users (email, password) VALUES ($1, $2) ON CONFLICT (email) DO NOTHING`
	res, err := s.db.Exec(query, email, passwordHash)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrUserAlreadyExists
	}
	return nil
}

func (s *UserStorage) GetPasswordHash(email string) (string, error) {
	var hash string

	query := `SELECT password FROM users WHERE email = $1`
	err := s.db.QueryRow(query, email).Scan(&hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrInvalidCredentials
		} else {
			return "", ErrDatabase
		}
	}
	return hash, nil
}
