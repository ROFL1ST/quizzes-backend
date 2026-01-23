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

# Stage 2: Runtime Environment (Python + Go)
# Gunakan image Python slim biar support library ML
FROM python:3.11-slim

WORKDIR /app

# Install dependencies sistem jika perlu (misal gcc buat build wheel)
# RUN apt-get update && apt-get install -y gcc

# 1. Setup ML Service (Python)
COPY ml-service/requirements.txt ./ml-service/requirements.txt
RUN pip install --no-cache-dir -r ml-service/requirements.txt

# Copy ML Service Code
COPY ml-service ./ml-service

# 2. Setup Backend (Go)
# Copy binary dari stage builder
COPY --from=builder /app/server .
# Copy .env & firebase credentials jika perlu (biasanya via Secret Manager di Cloud run)
COPY .env . 

# 3. Setup Script Startup
COPY start.sh .
RUN chmod +x start.sh

# Cloud Run Port
EXPOSE 8000
# ML Service Port (Internal)
EXPOSE 5002

# Jalankan script
CMD ["./start.sh"]