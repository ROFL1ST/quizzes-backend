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
# Copy .env
COPY .env . 

# 3. Setup Script Startup
COPY start.sh .
RUN chmod +x start.sh

# Cloud Run Port
EXPOSE 8000

# Jalankan script
CMD ["./start.sh"]