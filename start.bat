@echo off
chcp 65001 >nul
echo ========================================
echo   SmartSync — полный запуск
echo ========================================
echo.

:: 1. Собираем фронтенд
echo [1/3] Сборка фронтенда...
cd frontend
call npx vite build
if %errorlevel% neq 0 (
    echo Ошибка сборки фронтенда!
    pause
    exit /b %errorlevel%
)
cd ..

:: 2. Запускаем Docker (БД + бэкенд + фронт)
echo [2/3] Запуск Docker Compose...
docker compose up -d --build

if %errorlevel% neq 0 (
    echo Ошибка Docker Compose!
    pause
    exit /b %errorlevel%
)

:: 3. Готово
echo.
echo ========================================
echo   ✅ SmartSync запущен!
echo.
echo   Откройте: http://localhost
echo.
echo   API:      http://localhost:8000
echo   Grafana:  http://localhost:3000 (admin/admin)
echo   Prometheus: http://localhost:9090
echo ========================================
echo.
pause