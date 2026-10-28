package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// POSTGRES
	db, err := storage.NewPostgresDB(ctx, cfg.DBDSN)
	if err != nil {
		log.Fatalf("failed to init postgres: %v", err)
	}
	log.Println("Postgres инициализирован")
	defer db.Close()

	// REDIS
	rdb, err := storage.NewRedisDB(ctx, cfg.Redis)
	if err != nil {
		log.Fatalf("failed to init redis: %v", err)
	}
	log.Println("Redis инициализирован")
	defer rdb.Close()

	tokenStorage := storage.NewTokenStorage(rdb)
	userStorage := storage.NewUserStorage(db)
	authService := service.NewAuthService(userStorage, tokenStorage, []byte(cfg.JWTSecret))
	authHandler := handler.NewAuthHandler(authService)

	mux := http.NewServeMux()

	mux.HandleFunc("/{$}", frontendHandler)
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.RegisterHandler)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.LoginHandler)
	mux.Handle("POST /api/v1/orders", authHandler.AuthMiddleware(authHandler.UserRateLimitMiddleware(http.HandlerFunc(handler.OrdersHandler))))
	//mux.HandleFunc("POST /api/v1/delivery/select", handler.deliveryHandler)
	//mux.HandleFunc("GET /api/v1/delivery/status", handler.deliveryHandler)

	log.Printf("Сервер слушает на %s", cfg.Port)
	if err := http.ListenAndServe(cfg.Port, mux); err != nil {
		fmt.Println("Server error:", err)
	}
}
