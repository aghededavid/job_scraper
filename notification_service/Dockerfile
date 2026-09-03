# ── Build stage ──────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o notification_service ./cmd/main.go

# ── Development stage (hot-reload with Air) ────────────────────────
FROM golang:1.22-alpine AS development

WORKDIR /app
RUN apk add --no-cache git

# Install Air for hot-reload
RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

EXPOSE 8003
CMD ["air", "-c", ".air.toml"]

# ── Runtime stage ─────────────────────────────────────────────────
FROM alpine:3.19 AS production

WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/notification_service .

EXPOSE 8003
CMD ["./notification_service"]
