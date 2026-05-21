package main

import (
	"database/sql"
	"log"
	"os"
	"time" // Добавили time

	"smartsync/internal/auth/handler"
	"smartsync/internal/auth/repository"
	"smartsync/internal/auth/service"

	_ "github.com/lib/pq"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	dbURL := getEnv("DATABASE_URL", "postgres://user:password@127.0.0.1:5433/smartsync?sslmode=disable")

	var db *sql.DB // Изменили способ инициализации переменной
	var err error

	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Ошибка инициализации Postgres:", err)
	}
	defer db.Close()

	// ИССЛЕДОВАНИЕ: Цикл ожидания базы данных (Retry Pattern)
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		err = db.Ping()
		if err == nil {
			log.Println("✅ Успешное подключение к Postgres!")
			break
		}
		log.Printf("⚠️ Попытка %d: База данных недоступна, ждем 3 секунды... Ошибка: %v\n", i+1, err)
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		log.Fatal("❌ Критическая ошибка! Не удалось подключиться к базе данных после 5 попыток.")
	}

	repo := repository.NewAuthRepository(db)
	authService := service.NewAuthService(repo)
	httpHandler := handler.NewAuthHandler(authService)

	router := httpHandler.InitRoutes()

	log.Println("✅ Auth Service [JWT] запущен на порту 8081")
	router.Run(":8081")
}
