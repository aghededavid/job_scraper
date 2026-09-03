# Build stage
FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o remotescope ./cmd/api

# Development stage (with hot reload via Air)
FROM golang:1.22-alpine AS development
RUN apk add --no-cache git
RUN go install github.com/air-verse/air@latest
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
EXPOSE 8080
CMD ["air", "-c", ".air.toml"]

# Production stage
FROM scratch AS production
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/remotescope /remotescope
EXPOSE 8080
ENTRYPOINT ["/remotescope"]
