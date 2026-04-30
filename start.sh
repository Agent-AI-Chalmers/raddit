#!/bin/bash
set -e

echo "=== Starting Raddit ==="

# Backend
echo "[1/3] Downloading Go dependencies..."
cd backend
go mod tidy
echo "[2/3] Starting backend on :8080..."
go run . &
BACKEND_PID=$!
cd ..

# Frontend
echo "[3/3] Installing frontend dependencies..."
cd frontend
npm install --silent
echo "Starting frontend on :5173..."
npm run dev &
FRONTEND_PID=$!
cd ..

echo ""
echo "✓ Raddit is running!"
echo "  Frontend: http://localhost:5173"
echo "  Backend:  http://localhost:8080"
echo ""
echo "Press Ctrl+C to stop."

trap "kill $BACKEND_PID $FRONTEND_PID 2>/dev/null" SIGINT SIGTERM
wait
