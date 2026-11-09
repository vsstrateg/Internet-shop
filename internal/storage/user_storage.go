package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// UserStorage представляет хранилище для информации о пользователях
type UserStorage struct {
	db *sql.DB
}

var (
	// ErrUserAlreadyExists указывает, что пользователь с таким email уже существует
	ErrUserAlreadyExists = errors.New("пользователь с таким email уже существует")
	// ErrInvalidCredentials указывает, что предоставленные учетные данные недействительны
	ErrInvalidCredentials = errors.New("неверный логин или пароль")
	// ErrDatabase указывает на ошибку базы данных
	ErrDatabase = errors.New("ошибка базы данных")
)

// NewUserStorage создает новый экземпляр UserStorage
func NewUserStorage(db *sql.DB) *UserStorage {
	return &UserStorage{db: db}
}

// NewPostgresDB создает подключение к базе данных PostgreSQL
func NewPostgresDB(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("открытие pgx: %w", err)
	}

	db.SetMaxOpenConns(25)                 // максимум открытых соединений
	db.SetMaxIdleConns(25)                 // соединений в простое
	db.SetConnMaxLifetime(5 * time.Minute) // ротация соединений

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("пинг postgres: %w", err)
	}
	return db, nil
}

// SaveUser сохраняет нового пользователя в базе данных
func (s *UserStorage) SaveUser(ctx context.Context, email string, passwordHash []byte) error {
	query := `INSERT INTO users (email, password) VALUES ($1, $2) ON CONFLICT (email) DO NOTHING`
	res, err := s.db.ExecContext(ctx, query, email, passwordHash)
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

// GetPasswordHash извлекает хэш пароля пользователя по его email
func (s *UserStorage) GetPasswordHash(ctx context.Context, email string) (string, error) {
	var hash string

	query := `SELECT password FROM users WHERE email = $1`
	err := s.db.QueryRowContext(ctx, query, email).Scan(&hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrInvalidCredentials
		}
		return "", ErrDatabase
	}
	return hash, nil
}
