package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"github.com/sony/gobreaker" // Добавили библиотеку предохранителя
)

// Rate limiter: не более 100 запросов в секунду с одного IP
type ipRateLimiter struct {
	visitors sync.Map
	limit    int
	window   time.Duration
}

type visitor struct {
	count    int
	lastSeen time.Time
}

func newRateLimiter(limit int, window time.Duration) *ipRateLimiter {
	rl := &ipRateLimiter{
		limit:  limit,
		window: window,
	}
	// Фоновая горутина для очистки старых записей каждые 10 минут
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			rl.visitors.Range(func(key, value interface{}) bool {
				v := value.(*visitor)
				if time.Since(v.lastSeen) > rl.window {
					rl.visitors.Delete(key)
				}
				return true
			})
		}
	}()
	return rl
}

func (rl *ipRateLimiter) allow(ip string) bool {
	val, _ := rl.visitors.LoadOrStore(ip, &visitor{})
	v := val.(*visitor)
	v.lastSeen = time.Now()
	v.count++
	if v.count > rl.limit {
		return false
	}
	// Сбрасываем счётчик по истечении окна
	time.AfterFunc(rl.window, func() {
		v.count--
	})
	return true
}

var rateLimiter = newRateLimiter(100, 1*time.Second)

var jwtSecret []byte
var cb *gobreaker.CircuitBreaker

func init() {
	jwtSecret = []byte("smartsync_diploma_secret_key_2026")
	// Настройки предохранителя
	st := gobreaker.Settings{
		Name:        "Microservices-Gateway-CB",
		MaxRequests: 3,               // Сколько тестовых запросов пускать при проверке оживления
		Interval:    5 * time.Second, // Период сброса счетчиков
		Timeout:     7 * time.Second, // Как долго цепь остается разомкнутой
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Цепь размыкается, если было > 3 запросов и > 50% из них упали
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.5
		},
	}
	cb = gobreaker.NewCircuitBreaker(st)
}

// Настройка для WebSockets (разрешаем запросы с любых доменов)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	// JWT секрет из переменной окружения (может быть переопределён)
	if envSecret := os.Getenv("JWT_SECRET"); envSecret != "" {
		jwtSecret = []byte(envSecret)
	}

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		// Rate limiting по IP
		ip := c.ClientIP()
		if !rateLimiter.allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Слишком много запросов. Попробуйте позже."})
			return
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	natsURL := getEnv("NATS_URL", "nats://localhost:4222")
	authURL := getEnv("AUTH_SERVICE_URL", "http://localhost:8081")
	taskURL := getEnv("TASK_SERVICE_URL", "http://localhost:8080")
	auditURL := getEnv("AUDIT_SERVICE_URL", "http://localhost:8083")

	// Подключаемся к NATS
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Println("⚠️ ВНИМАНИЕ: NATS недоступен. WebSockets работать не будут.")
	} else {
		defer nc.Close()
		log.Println("✅ Gateway подключен к NATS для трансляции событий")
	}

	// Эндпоинт для WebSockets
	r.GET("/ws", func(c *gin.Context) {
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("❌ Ошибка Upgrade WS:", err)
			return
		}
		defer ws.Close()

		if nc == nil {
			log.Println("❌ Ошибка WS: NATS не подключен!")
			return
		}

		log.Println("🟢 Клиент подключился к WebSocket!")

		// ИССЛЕДОВАНИЕ: Используем АСИНХРОННУЮ подписку (Event-Driven)
		sub, err := nc.Subscribe("project.updated", func(msg *nats.Msg) {
			log.Printf("📨 NATS поймал событие! Пушим в браузер: %s\n", string(msg.Data))
			err := ws.WriteMessage(websocket.TextMessage, msg.Data)
			if err != nil {
				log.Println("❌ Ошибка отправки в WS:", err)
			}
		})

		if err != nil {
			log.Println("❌ Ошибка подписки NATS:", err)
			return
		}
		defer sub.Unsubscribe()

		// Удерживаем соединение открытым (читаем системные пинги)
		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				log.Println("🔴 Клиент отключился от WebSocket")
				break
			}
		}
	})

	// Динамические прокси вместо жесткого localhost
	authProxy := reverseProxy(authURL)
	r.POST("/register", authProxy)
	r.POST("/login", authProxy)

	taskProxy := reverseProxy(taskURL)
	auditProxy := reverseProxy(auditURL)

	protected := r.Group("/")
	protected.Use(authMiddleware())
	{
		protected.POST("/tasks", taskProxy)
		protected.GET("/tasks/:id", taskProxy)
		protected.PUT("/tasks/:id", taskProxy)
		protected.DELETE("/tasks/:id", taskProxy)
		protected.POST("/tasks/:id/dependencies", taskProxy)
		protected.DELETE("/tasks/:id/dependencies/:dep_id", taskProxy)
		protected.PATCH("/tasks/:id/status", taskProxy)

		protected.GET("/user/profile", authProxy)
		protected.PUT("/user/profile", authProxy)
		protected.GET("/user/audit", auditProxy)

		protected.GET("/invitations/my", taskProxy)
		protected.GET("/projects/:project_id/members", taskProxy)

		protected.GET("/projects", taskProxy)
		protected.POST("/projects", taskProxy)
		protected.DELETE("/projects/:project_id", taskProxy)
		protected.PUT("/projects/:project_id", taskProxy)
		protected.POST("/projects/:project_id/members", taskProxy)
		protected.DELETE("/projects/:project_id/dependencies", taskProxy)
		protected.GET("/projects/:project_id/graph", taskProxy)
		protected.GET("/projects/:project_id/tasks", taskProxy)

		protected.DELETE("/projects/:project_id/members/:user_id", taskProxy)
		protected.PATCH("/projects/:project_id/members/:user_id", taskProxy)

		protected.GET("/logs/:task_id", auditProxy)
		protected.POST("/internal/users/bulk", authProxy)
		protected.GET("/users/:id", authProxy)
		protected.POST("/tasks/:id/comments", taskProxy)
		protected.GET("/tasks/:id/comments", taskProxy)
		protected.GET("/projects/:project_id/stats", taskProxy)
		protected.GET("/projects/:project_id/milestones", taskProxy)
		protected.POST("/projects/:project_id/milestones", taskProxy)
	}

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":8000",
		Handler: r,
	}

	go func() {
		log.Println("API Gateway запущен на порту 8000")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка Gateway: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Gateway завершает работу...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Ошибка при остановке Gateway: %v", err)
	}
	if nc != nil {
		nc.Drain()
	}
	log.Println("Gateway остановлен")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
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

		claims, _ := token.Claims.(jwt.MapClaims)
		userID := fmt.Sprintf("%v", claims["user_id"])
		c.Request.Header.Set("X-User-ID", userID)
		c.Next()
	}
}

func reverseProxy(target string) gin.HandlerFunc {
	targetURL, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Перехватываем системные ошибки прокси (например, сервис физически выключен)
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		rw.WriteHeader(http.StatusBadGateway) // 502 код
		rw.Write([]byte(fmt.Sprintf(`{"error": "Внутренний сервис %s недоступен"}`, targetURL.Host)))
	}

	return func(c *gin.Context) {
		// Пропускаем запрос через предохранитель
		_, err := cb.Execute(func() (interface{}, error) {

			proxy.ServeHTTP(c.Writer, c.Request)

			// Если сервис вернул статус 5xx, считаем это поломкой микросервиса
			if c.Writer.Status() >= http.StatusInternalServerError {
				return nil, fmt.Errorf("микросервис вернул ошибку сервера")
			}
			return nil, nil
		})

		// Обработка состояния разомкнутой цепи
		if err != nil {
			if err == gobreaker.ErrOpenState {
				// Цепь разомкнута: рубим запрос сразу, не нагружая упавший сервис!
				c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
					"error": "Система перегружена. Включился предохранитель (Circuit Breaker). Подождите 7 секунд.",
					"state": "OPEN",
				})
			}
		}
	}
}
