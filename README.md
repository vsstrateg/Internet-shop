# 🛒 E-Commerce Backend Service

Backend-сервис для интернет-магазина, разработанный на Go с соблюдением принципов чистой архитектуры, поддержкой авторизации через JWT, кэшированием и Rate Limiting в Redis, а также автоматическими миграциями в PostgreSQL.

---

### 🛠 Стек технологий

![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-18-316192?style=for-the-badge&logo=postgresql&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=for-the-badge&logo=redis&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![JWT](https://img.shields.io/badge/JWT-Auth-000000?style=for-the-badge&logo=jsonwebtokens&logoColor=white)

* **Язык**: Go (стандартная библиотека `net/http`)
* **База данных**: PostgreSQL 18 (драйвер `pgx/v5`, пул соединений `database/sql`)
* **Кэширование & Rate Limiting**: Redis (`go-redis/v9`)
* **Миграции**: `golang-migrate` (автоматический накат версий при старте приложения)
* **Безопасность**: `golang-jwt/jwt/v5`, `golang.org/x/crypto/bcrypt`
* **Инфраструктура**: Docker, Docker Compose

---

### 🚀 Архитектурные особенности и ключевые фичи

* **Clean Architecture**: 
  * Строгое разделение ответственности на 3 слоя: `handler` (транспорт) $\rightarrow$ `service` (бизнес-логика) $\rightarrow$ `storage` (доступ к данным).
  * Инверсия зависимостей (**Dependency Injection**): слои связаны через интерфейсы, объявляемые на стороне потребителя.
* **Аутентификация и безопасность**:
  * Хэширование паролей с солью с помощью **bcrypt**.
  * Работа с сессиями через **JWT**, сохраняемый в защищенные **HttpOnly Cookies** (защита от XSS-атак).
  * Валидация активных токенов через **Redis** с возможностью мгновенной инвалидации сессий.
* **Rate Limiting (Защита от спама и DDoS)**:
  * Ограничение частоты создания заказов для каждого пользователя с использованием счетчиков в Redis (скользящее окно).
* **Защита данных**:
  * Проверка прав доступа: пользователи могут просматривать и оформлять доставку только для своих собственных заказов.
* **Отказоустойчивость инфраструктуры**:
  * Сервис запускается через Docker Compose только после прохождения `healthcheck` со стороны PostgreSQL.
  * Пул соединений PostgreSQL тонко настроен под нагрузку (`SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`).

---

### 📁 Структура проекта

```text
.
├── cmd/
│   └── app/
│       └── main.go              # Точка входа, DI и сборка зависимостей
├── internal/
│   ├── config/                  # Загрузка конфигурации из переменных окружения
│   ├── handler/                 # HTTP-хендлеры, middleware, роутинг
│   ├── service/                 # Бизнес-логика (Auth, Orders, Delivery, JWT)
│   ├── storage/                 # Слой работы с базами данных (Postgres, Redis)
│   └── migrations/              # SQL-миграции схемы БД (.up.sql / .down.sql)
├── docker-compose.yml           # Оркестрация сервисов (App, Postgres, Redis)
├── Dockerfile                   # Описание сборки контейнера Go-приложения
├── index.html                   # Демо-фронтенд для проверки функционала
└── go.mod
```

---

### 🚦 Быстрый запуск

Для запуска требуется только установленный **Docker** и **Docker Compose**.

1. **Клонируйте репозиторий:**
   ```bash
   git clone https://github.com/vsstrateg/Internet-shop.git
   cd Internet-shop
   ```

2. **Создайте файл окружения из примера:**
   ```bash
   cp .env.example .env
   ```

3. **Запустите проект одной командой:**
   ```bash
   docker compose up --build
   ```

4. **Готово!** Приложение будет доступно в браузере по адресу:
   👉 **[http://localhost:8080](http://localhost:8080)**

---

### 📡 API Спецификация

| Метод | Эндпоинт | Доступ | Описание |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/v1/auth/register` | Публичный | Регистрация нового пользователя |
| `POST` | `/api/v1/auth/login` | Публичный | Авторизация и выдача JWT в Cookie |
| `POST` | `/api/v1/orders` | Авторизован + RateLimit | Создание нового заказа |
| `POST` | `/api/v1/delivery/select` | Авторизован | Выбор службы доставки (CDEK, Yandex, Boxberry) |
| `GET` | `/api/v1/delivery/status` | Авторизован | Проверка статуса доставки заказа |
| `GET` | `/` | Публичный | Веб-интерфейс интернет-магазина |