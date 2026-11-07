FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/app

FROM alpine:3.24
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/index.html .
COPY --from=builder /app/internal/migrations ./internal/migrations

CMD ["./server"]