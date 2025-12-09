# Stage 1: Build aplikasi
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy file dependency & download library
COPY go.mod go.sum ./
RUN go mod download

# Copy seluruh source code
COPY . .

# Build aplikasi jadi binary bernama 'server'
RUN go build -o server main.go

# Stage 2: Image untuk dijalankan (lebih ringan)
FROM alpine:latest

WORKDIR /app

# Copy binary dari stage builder
COPY --from=builder /app/server .

# Cloud Run butuh ini biar tau port berapa yang dibuka
EXPOSE 8000

# Jalankan aplikasi
CMD ["./server"]