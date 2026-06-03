Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  SmartSync — полный запуск" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 1. Собираем фронтенд
Write-Host "[1/3] Сборка фронтенда..." -ForegroundColor Yellow
Set-Location frontend
npx vite build
if ($LASTEXITCODE -ne 0) {
    Write-Host "Ошибка сборки фронтенда!" -ForegroundColor Red
    pause
    exit $LASTEXITCODE
}
Set-Location ..

# 2. Запускаем Docker (БД + бэкенд + фронт)
Write-Host "[2/3] Запуск Docker Compose..." -ForegroundColor Yellow
docker compose up -d --build

if ($LASTEXITCODE -ne 0) {
    Write-Host "Ошибка Docker Compose!" -ForegroundColor Red
    pause
    exit $LASTEXITCODE
}

# 3. Готово
Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "  ✅ SmartSync запущен!" -ForegroundColor Green
Write-Host ""
Write-Host "  Откройте: http://localhost" -ForegroundColor White
Write-Host ""
Write-Host "  API:       http://localhost:8000" -ForegroundColor Gray
Write-Host "  Grafana:   http://localhost:3000 (admin/admin)" -ForegroundColor Gray
Write-Host "  Prometheus: http://localhost:9090" -ForegroundColor Gray
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
pause