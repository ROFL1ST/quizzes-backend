#!/bin/sh

# Jalankan Service ML di Background (Port 5002)
echo "Starting ML Service..."
python ml-service/main.py &

# Simpan PID ML Service
ML_PID=$!

# Tunggu sebentar
sleep 2

# Jalankan Service Utama (Go) di Foreground (Port 8000/Default)
echo "Starting Go Server..."
./server &
SERVER_PID=$!

# Trap SIGTERM/SIGINT untuk matikan keduanya
trap "kill $ML_PID $SERVER_PID" TERM INT

# Wait agar container tidak exit
wait $SERVER_PID
