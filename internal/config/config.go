package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	DBDSN     string
	JWTSecret string
	Redis     string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Предупреждение: .env файл не найден, читаем переменные системы")
	}
	return &Config{
		Port:      getEnv("SERVER_PORT", ":8080"),
		DBDSN:     getEnv("DB_DSN", ""),
		JWTSecret: getEnv("JWT_SECRET", ""),
		Redis:     getEnv("REDIS_ADDR", ""),
	}
}

// Вспомогательная функция: берет значение из ОС или возвращает default
func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
