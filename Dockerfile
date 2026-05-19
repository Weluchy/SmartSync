# Этап 1: Сборка (Builder)
# Используем официальный образ Go
FROM golang:alpine AS builder
# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы зависимостей и скачиваем их
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Получаем имя сервиса через аргумент сборки
ARG SERVICE_NAME

# Собираем бинарник конкретного сервиса. 
# Флаги -ldflags="-s -w" удаляют отладочную информацию, делая файл еще меньше
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o microservice ./cmd/${SERVICE_NAME}/main.go

# Этап 2: Финальный минималистичный образ
# Используем пустой образ alpine, чтобы контейнер весил 10-15 МБ, а не 1 ГБ
FROM alpine:latest

WORKDIR /app

# Копируем готовый бинарник из первого этапа
COPY --from=builder /app/microservice .

# Запускаем наш микросервис
CMD ["./microservice"]