package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("smartsync_diploma_secret_key_2026")

func main() {
	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 1. Маршруты авторизации (прокидываем на порт 8081 БЕЗ проверки токена)
	authProxy := reverseProxy("http://localhost:8081")
	r.POST("/register", authProxy)
	r.POST("/login", authProxy)

	// 2. Маршруты задач (прокидываем на порт 8080 С ПРОВЕРКОЙ токена)
	taskProxy := reverseProxy("http://localhost:8080")
	protected := r.Group("/")
	protected.Use(authMiddleware())
	{
		protected.POST("/tasks", taskProxy)
		protected.POST("/tasks/:id/dependencies", taskProxy)
		protected.DELETE("/dependencies", taskProxy)
		protected.GET("/graph", taskProxy)
	}

	log.Println("API Gateway запущен на порту 8000")
	r.Run(":8000")
}

// Middleware для проверки JWT на уровне шлюза
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Неверный токен"})
			return
		}

		// Если токен ок, Gateway пропускает запрос дальше
		c.Next()
	}
}

func reverseProxy(target string) gin.HandlerFunc {
	url, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(url)
	return func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
