# Этап 1: Сборка (Builder)
# Используем официальный образ Go с поддержкой multi-arch
# Передавать --build-arg TARGETARCH=arm64 (для Oracle Ampere) или amd64
ARG TARGETARCH=amd64
ARG TARGETOS=linux

FROM --platform=$BUILDPLATFORM golang:alpine AS builder

# Устанавливаем утилиты для сборки под другую архитектуру
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Копируем файлы зависимостей и скачиваем их
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Получаем имя сервиса через аргумент сборки
ARG SERVICE_NAME

# Собираем бинарник конкретного сервиса с учётом целевой архитектуры.
# Флаги -ldflags="-s -w" удаляют отладочную информацию
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o microservice ./cmd/${SERVICE_NAME}/main.go

# Этап 2: Финальный минималистичный образ
FROM --platform=linux/${TARGETARCH} alpine:latest

WORKDIR /app

# Копируем готовый бинарник из первого этапа
COPY --from=builder /app/microservice .

# Запускаем наш микросервис
CMD ["./microservice"]