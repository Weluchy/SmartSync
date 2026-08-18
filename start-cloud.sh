#!/bin/bash
# ============================================================
# SmartSync — Скрипт запуска на облачном сервере
# Использование:
#   chmod +x start-cloud.sh
#   ./start-cloud.sh
# ============================================================

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}============================================${NC}"
echo -e "${BLUE}  SmartSync — Deploy to Cloud${NC}"
echo -e "${BLUE}============================================${NC}"

# Проверка наличия .env.cloud
if [ ! -f ".env.cloud" ]; then
    echo -e "${RED}❌ Файл .env.cloud не найден!${NC}"
    echo -e "${YELLOW}Скопируйте .env.cloud в .env.cloud и заполните свои данные:${NC}"
    echo "  cp .env.cloud .env.cloud.prod"
    echo "  # отредактируйте .env.cloud.prod"
    echo "  mv .env.cloud.prod .env.cloud"
    exit 1
fi

# Загружаем переменные из .env.cloud
export $(grep -v '^\s*#' .env.cloud | grep -v '^\s*$' | xargs)

# Определяем архитектуру сервера
ARCH=$(uname -m)
case $ARCH in
    aarch64|arm64)
        TARGETARCH="arm64"
        echo -e "${GREEN}✅ Обнаружена архитектура: ARM64 (${ARCH})${NC}"
        ;;
    x86_64|amd64)
        TARGETARCH="amd64"
        echo -e "${GREEN}✅ Обнаружена архитектура: AMD64 (${ARCH})${NC}"
        ;;
    *)
        echo -e "${RED}❌ Неизвестная архитектура: ${ARCH}${NC}"
        exit 1
        ;;
esac

# Экспортируем архитектуру для docker-compose
export TARGETARCH

# Проверка Docker
if ! command -v docker &> /dev/null; then
    echo -e "${YELLOW}⚠️ Docker не найден. Устанавливаем...${NC}"
    curl -fsSL https://get.docker.com -o get-docker.sh
    sudo sh get-docker.sh
    sudo usermod -aG docker $USER
    echo -e "${GREEN}✅ Docker установлен. Перезапустите сессию: newgrp docker${NC}"
    exit 0
fi

# Проверка Docker Compose
if ! docker compose version &> /dev/null; then
    echo -e "${YELLOW}⚠️ Docker Compose не найден. Устанавливаем...${NC}"
    sudo apt-get update && sudo apt-get install -y docker-compose-plugin
fi

echo -e "${GREEN}🐳 Docker: $(docker --version)${NC}"
echo -e "${GREEN}🐳 Docker Compose: $(docker compose version)${NC}"

# Останавливаем старые контейнеры (если есть)
echo -e "${YELLOW}🔄 Останавливаем старые контейнеры...${NC}"
docker compose -f docker-compose.cloud.yml down 2>/dev/null || true

# Собираем образы с учётом архитектуры
echo -e "${BLUE}🔨 Сборка образов (архитектура: ${TARGETARCH})...${NC}"
docker compose -f docker-compose.cloud.yml build --build-arg TARGETARCH=${TARGETARCH}

# Запускаем
echo -e "${BLUE}🚀 Запуск контейнеров...${NC}"
docker compose -f docker-compose.cloud.yml up -d

# Проверка статуса
echo -e "\n${GREEN}📊 Статус контейнеров:${NC}"
docker compose -f docker-compose.cloud.yml ps

# Проверка, что всё работает
echo -e "\n${GREEN}🔍 Проверка сервисов:${NC}"

check_service() {
    local name=$1
    local url=$2
    if curl -sf "$url" > /dev/null 2>&1; then
        echo -e "${GREEN}  ✅ $name ($url) — доступен${NC}"
    else
        echo -e "${YELLOW}  ⚠️  $name ($url) — не отвечает (возможно, ещё грузится)${NC}"
    fi
}

sleep 3
check_service "Gateway" "http://localhost:8000"
check_service "Auth Service" "http://localhost:8081"
check_service "Task Service" "http://localhost:8080"
check_service "Prometheus" "http://localhost:9090"
check_service "Grafana" "http://localhost:3000"

echo -e "\n${GREEN}============================================${NC}"
echo -e "${GREEN}  SmartSync запущен!${NC}"
echo -e "${GREEN}============================================${NC}"
echo -e ""
echo -e "${BLUE}Сервисы доступны:${NC}"
echo -e "  Gateway API:    http://localhost:8000"
echo -e "  Grafana:        http://localhost:3000 (admin/${GF_SECURITY_ADMIN_PASSWORD:-admin})"
echo -e "  Prometheus:     http://localhost:9090"
echo -e "  NATS:           port 4222"
echo -e ""
echo -e "${YELLOW}Для просмотра логов:${NC}"
echo -e "  docker compose -f docker-compose.cloud.yml logs -f"
echo -e ""
echo -e "${YELLOW}Для остановки:${NC}"
echo -e "  docker compose -f docker-compose.cloud.yml down"