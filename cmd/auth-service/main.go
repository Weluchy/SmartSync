package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	var db *sql.DB
	var err error

	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Ошибка инициализации Postgres:", err)
	}
	defer db.Close()

	// Retry Pattern: цикл ожидания базы данных
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

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":8081",
		Handler: router,
	}

	go func() {
		log.Println("✅ Auth Service [JWT] запущен на порту 8081")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка Auth Service: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Auth Service завершает работу...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("Auth Service остановлен")
}
