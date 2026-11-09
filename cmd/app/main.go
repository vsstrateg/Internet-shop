// Package main содержит точку входа для запуска сервера аутентификации и управления заказами
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"auth-service/internal/config"
	"auth-service/internal/handler"
	"auth-service/internal/service"
	"auth-service/internal/storage"
)

func frontendHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./index.html")
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := run(); err != nil {
		slog.Error("Ошибка при запуске сервера", "error", err)
		os.Exit(1)
	}
}

func run() error {

	cfg := config.LoadConfig()

	initCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	// POSTGRES
	db, err := storage.NewPostgresDB(initCtx, cfg.DBDSN)
	if err != nil {
		return fmt.Errorf("ошибка инициализации Postgres: %w", err)
	}
	slog.Info("Postgres инициализирован")
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("ошибка закрытия БД", "error", err)
		}
	}()

	if err := storage.RunMigrations(db); err != nil {
		return fmt.Errorf("ошибка миграции БД: %w", err)
	}

	// REDIS
	rdb, err := storage.NewRedisDB(initCtx, cfg.Redis)
	if err != nil {
		return fmt.Errorf("ошибка инициализации Redis: %w", err)
	}
	slog.Info("Redis инициализирован")
	defer func() {
		if err := rdb.Close(); err != nil {
			slog.Error("ошибка закрытия Redis", "error", err)
		}
	}()

	cancel()

	tokenStorage := storage.NewTokenStorage(rdb)
	userStorage := storage.NewUserStorage(db)
	orderStorage := storage.NewOrderStorage(db)
	deliveryStorage := storage.NewDeliveryStorage(db)

	authService := service.NewAuthService(userStorage, tokenStorage, []byte(cfg.JWTSecret))
	orderService := service.NewOrderService(orderStorage)
	deliveryService := service.NewDeliveryService(deliveryStorage)

	authHandler := handler.NewAuthHandler(authService)
	orderHandler := handler.NewOrderHandler(orderService)
	deliveryHandler := handler.NewDeliveryHandler(deliveryService)

	mux := http.NewServeMux()

	mux.HandleFunc("/{$}", frontendHandler)
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.RegisterHandler)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.LoginHandler)
	mux.Handle("POST /api/v1/orders", authHandler.AuthMiddleware(authHandler.UserRateLimitMiddleware(http.HandlerFunc(orderHandler.CreateOrderHandler))))
	mux.Handle("POST /api/v1/delivery/select", authHandler.AuthMiddleware(http.HandlerFunc(deliveryHandler.SelectDeliveryHandler)))
	mux.Handle("GET /api/v1/delivery/status", authHandler.AuthMiddleware(http.HandlerFunc(deliveryHandler.CheckStatusDeliveryHandler)))

	srv := &http.Server{
		Addr:              cfg.Port,
		Handler:           mux,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Ошибка сервера", "error", err)
		}
	}()
	slog.Info("Сервер слушает", "port", cfg.Port)

	stopCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-stopCtx.Done()
	slog.Info("Получен сигнал остановки, завершаем работу...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Ошибка при принудительной остановке сервера", "error", err)
	}
	slog.Info("Сервер успешно и плавно остановлен")

	return nil
}
