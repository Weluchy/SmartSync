package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware проверяет наличие и валидность JWT токена
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Доступ запрещен. Отсутствует токен."})
			c.Abort()
			return
		}

		// Убираем слово "Bearer " из заголовка
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte("smartsync_diploma_secret_key_2026"), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Токен недействителен или просрочен"})
			c.Abort()
			return
		}

		// Достаем user_id из расшифрованного токена и кладем в контекст запроса
		claims, _ := token.Claims.(jwt.MapClaims)
		c.Set("user_id", int(claims["user_id"].(float64)))
		c.Next()
	}
}
