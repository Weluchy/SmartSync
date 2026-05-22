package main

import (
	"database/sql"
	"log"
	"os"
	"strconv"
	"time" // Добавили time

	"smartsync/internal/auth/handler"
	"smartsync/internal/auth/repository"
	"smartsync/internal/auth/service"

	"github.com/gin-gonic/gin"
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
	// ФИКС: Эндпоинт для перевода ID в Имена для Истории
	router.POST("/internal/users/bulk", func(c *gin.Context) {
		var req struct {
			IDs []int `json:"ids"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "bad request"})
			return
		}
		result := make(map[string]string)
		for _, id := range req.IDs {
			var username string
			// Прямой запрос в БД для скорости
			err := db.QueryRow("SELECT username FROM users WHERE id = $1", id).Scan(&username)
			if err == nil {
				result[strconv.Itoa(id)] = username
			}
		}
		c.JSON(200, result)
	})
	log.Println("✅ Auth Service [JWT] запущен на порту 8081")
	router.Run(":8081")
}
