package main

import (
	"context"
	"errors"
	"log"
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

	cfg := config.LoadConfig()

	initCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// POSTGRES
	db, err := storage.NewPostgresDB(initCtx, cfg.DBDSN)
	if err != nil {
		log.Fatalf("failed to init postgres: %v", err)
	}
	log.Println("Postgres инициализирован")
	defer db.Close()

	if err := storage.RunMigrations(db); err != nil {
		log.Fatalf("Ошибка миграции БД: %v", err)
	}

	// REDIS
	rdb, err := storage.NewRedisDB(initCtx, cfg.Redis)
	if err != nil {
		log.Fatalf("failed to init redis: %v", err)
	}
	log.Println("Redis инициализирован")
	defer rdb.Close()

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
			log.Fatalf("Ошибка сервера: %v", err)
		}
	}()
	log.Printf("Сервер слушает на %s", cfg.Port)

	stopCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-stopCtx.Done()
	log.Println("Получен сигнал остановки, завершаем работу...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Ошибка при принудительной остановке сервера: %v", err)
	}
	log.Println("Сервер успешно и плавно остановлен")
}
