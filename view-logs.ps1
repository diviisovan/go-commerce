# PowerShell script to view runtime logs
# Usage: .\view-logs.ps1

Write-Host "=== eCommerce Server Logs ===" -ForegroundColor Green
Write-Host "Press Ctrl+C to stop viewing logs`n" -ForegroundColor Yellow

if (Test-Path "logs\app.log") {
    Get-Content "logs\app.log" -Wait -Tail 50
} else {
    Write-Host "Log file not found. Make sure the server is running." -ForegroundColor Red
    Write-Host "Start the server with: go run ./cmd/server" -ForegroundColor Yellow
}
