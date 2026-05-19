package main

import (
	"database/sql"
	"log"
	"os" // Добавили для чтения переменных окружения

	"smartsync/internal/auth/handler"
	"smartsync/internal/auth/repository"
	"smartsync/internal/auth/service"

	_ "github.com/lib/pq"
)

// Вспомогательная функция для переменных окружения
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	// Читаем URL базы данных из окружения Docker
	dbURL := getEnv("DATABASE_URL", "postgres://user:password@127.0.0.1:5433/smartsync?sslmode=disable")

	// Подключение к БД
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Ошибка инициализации Postgres:", err)
	}
	defer db.Close()

	// ИССЛЕДОВАНИЕ: Жестко проверяем, что база реально доступна
	if err := db.Ping(); err != nil {
		log.Fatal("Ошибка! База данных недоступна:", err)
	}

	// Сборка слоев
	repo := repository.NewAuthRepository(db)
	authService := service.NewAuthService(repo)
	httpHandler := handler.NewAuthHandler(authService)

	router := httpHandler.InitRoutes()

	log.Println("Auth Service [JWT] запущен на порту 8081")
	router.Run(":8081")
}
