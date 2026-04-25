package main

import (
	"database/sql"
	"log"

	"smartsync/internal/auth/handler"
	"smartsync/internal/auth/repository"
	"smartsync/internal/auth/service"

	_ "github.com/lib/pq"
)

func main() {
	// Подключение к БД
	db, err := sql.Open("postgres", "postgres://user:password@127.0.0.1:5433/smartsync?sslmode=disable")
	if err != nil {
		log.Fatal("Postgres error:", err)
	}
	defer db.Close()

	// Сборка слоев
	repo := repository.NewAuthRepository(db)
	authService := service.NewAuthService(repo)
	httpHandler := handler.NewAuthHandler(authService)

	router := httpHandler.InitRoutes()

	log.Println("Auth Service [JWT] запущен на порту 8081")
	router.Run(":8081")
}
