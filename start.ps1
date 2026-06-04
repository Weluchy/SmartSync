Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  SmartSync — запуск микросервисов" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

docker compose down -v
docker compose up -d --build

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "  ✅ SmartSync запущен!" -ForegroundColor Green
Write-Host "  Сайт:      http://localhost" -ForegroundColor White
Write-Host "  Grafana:   http://localhost:3000" -ForegroundColor Gray
Write-Host "========================================" -ForegroundColor Green