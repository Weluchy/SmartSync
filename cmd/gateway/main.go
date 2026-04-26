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

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	authProxy := reverseProxy("http://localhost:8081")
	r.POST("/register", authProxy)
	r.POST("/login", authProxy)

	taskProxy := reverseProxy("http://localhost:8080")
	protected := r.Group("/")
	protected.Use(authMiddleware())
	{
		protected.POST("/tasks", taskProxy)
		protected.PUT("/tasks/:id", taskProxy)
		protected.DELETE("/tasks/:id", taskProxy)
		protected.POST("/tasks/:id/dependencies", taskProxy)
		protected.DELETE("/tasks/:id/dependencies/:dep_id", taskProxy)
		protected.PATCH("/tasks/:id/status", taskProxy)

		protected.GET("/projects", taskProxy)
		protected.POST("/projects", taskProxy)
		protected.DELETE("/projects/:project_id", taskProxy)
		protected.PUT("/projects/:project_id", taskProxy)
		protected.POST("/projects/:project_id/members", taskProxy)
		protected.DELETE("/projects/:project_id/dependencies", taskProxy)
		protected.GET("/projects/:project_id/graph", taskProxy)
		protected.GET("/projects/:project_id/tasks", taskProxy)
	}

	log.Println("API Gateway запущен на порту 8000")
	r.Run(":8000")
}

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
