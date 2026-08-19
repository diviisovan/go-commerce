#!/bin/bash
# Shell script to view runtime logs
# Usage: ./view-logs.sh

echo "=== eCommerce Server Logs ==="
echo "Press Ctrl+C to stop viewing logs"
echo ""

if [ -f "logs/app.log" ]; then
    tail -f logs/app.log
else
    echo "Log file not found. Make sure the server is running."
    echo "Start the server with: go run ./cmd/server"
fi
