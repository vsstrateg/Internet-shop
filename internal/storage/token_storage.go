package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrTokenNotFound = errors.New("Токена не найдено")
)

type TokenStorage struct {
	rdb *redis.Client
}

func NewTokenStorage(rdb *redis.Client) *TokenStorage {
	return &TokenStorage{
		rdb: rdb,
	}
}

func NewRedisDB(ctx context.Context, addr string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     "",
		DB:           0,
		Protocol:     2,
		PoolSize:     50, // размер пула под нагрузку
		MinIdleConns: 10,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return rdb, nil
}

func (s *TokenStorage) SaveToken(ctx context.Context, email string, token string, ttl time.Duration) error {
	key := "user:" + email

	return s.rdb.Set(ctx, key, token, ttl).Err()
}

func (s *TokenStorage) GetToken(ctx context.Context, email string) (string, error) {
	key := "user:" + email
	val, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrTokenNotFound
		}
		return "", err
	}
	return val, nil
}

func (s *TokenStorage) IncrementRequests(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	size, err := s.rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	if size == 1 {
		if err := s.rdb.Expire(ctx, key, ttl).Err(); err != nil {
			return 0, err
		}
	}

	return size, nil
}
