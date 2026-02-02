# Stage 1: Build aplikasi Go
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN go build -o server main.go

# Stage 2: Runtime Environment (Pure Go)
FROM alpine:latest

WORKDIR /app

# Install dependencies if needed (e.g., for health checks or basic tools)
RUN apk --no-cache add ca-certificates bash curl

# 2. Setup Backend (Go)
# Copy binary dari stage builder
COPY --from=builder /app/server .
# COPY .env . # Tidak perlu copy .env untuk Cloud Run (gunakan Environment Variables)

# Cloud Run Port
EXPOSE 8080

# Jalankan Service
CMD ["./server"]