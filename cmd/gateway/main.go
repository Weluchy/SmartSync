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
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"github.com/sony/gobreaker"
)

// не более 100 запросов в сек с одного ip
type ipRateLimiter struct {
	visitors sync.Map
	limit    int64
	window   time.Duration
	mu       sync.Mutex
}

type visitor struct {
	count    int64
	lastSeen time.Time
}

func newRateLimiter(limit int, window time.Duration) *ipRateLimiter {
	rl := &ipRateLimiter{
		limit:  int64(limit),
		window: window,
	}
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			rl.visitors.Range(func(key, value interface{}) bool {
				v := value.(*visitor)
				if time.Since(v.lastSeen) > rl.window*2 {
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
	newCount := atomic.AddInt64(&v.count, 1)
	if newCount > rl.limit {
		return false
	}
	if newCount == 1 {
		go func() {
			time.Sleep(rl.window)
			atomic.AddInt64(&v.count, -newCount) // сбрасываем накопленное
		}()
	}
	return true
}

var rateLimiter = newRateLimiter(100, 1*time.Second)

var jwtSecret []byte

// отдельный cb на сервис, чтобы один сбой не валил всё
var taskCB *gobreaker.CircuitBreaker
var authCB *gobreaker.CircuitBreaker
var auditCB *gobreaker.CircuitBreaker

func newCB(name string) *gobreaker.CircuitBreaker {
	st := gobreaker.Settings{
		Name:        name,
		MaxRequests: 5, // больше при проверке тестов
		Interval:    10 * time.Second,
		Timeout:     5 * time.Second, // Быстрее восстанавливаемся
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 5 && failureRatio >= 0.6
		},
	}
	return gobreaker.NewCircuitBreaker(st)
}

func init() {
	jwtSecret = []byte("smartsync_diploma_secret_key_2026")
	taskCB = newCB("Task-Service-CB")
	authCB = newCB("Auth-Service-CB")
	auditCB = newCB("Audit-Service-CB")
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	if envSecret := os.Getenv("JWT_SECRET"); envSecret != "" {
		jwtSecret = []byte(envSecret)
	}

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

	natsURL := getEnv("NATS_URL", "nats://localhost:4222")
	authURL := getEnv("AUTH_SERVICE_URL", "http://localhost:8081")
	taskURL := getEnv("TASK_SERVICE_URL", "http://localhost:8080")
	auditURL := getEnv("AUDIT_SERVICE_URL", "http://localhost:8083")

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Println("⚠️ ВНИМАНИЕ: NATS недоступен. WebSockets работать не будут.")
	} else {
		defer nc.Close()
		log.Println("✅ Gateway подключен к NATS для трансляции событий")
	}

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

		// форвардим обновления проекта в браузер
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

		// держим соединение открытым
		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				log.Println("🔴 Клиент отключился от WebSocket")
				break
			}
		}
	})

	authProxy := reverseProxy(authURL, authCB)
	taskProxy := reverseProxy(taskURL, taskCB)
	auditProxy := reverseProxy(auditURL, auditCB)

	r.POST("/register", authProxy)
	r.POST("/login", authProxy)

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
		protected.POST("/projects/:project_id/archive", taskProxy)
		protected.POST("/projects/:project_id/unarchive", taskProxy)

		protected.GET("/logs/:task_id", auditProxy)
		protected.POST("/internal/users/bulk", authProxy)
		protected.GET("/users/:id", authProxy)
		protected.POST("/tasks/:id/comments", taskProxy)
		protected.GET("/tasks/:id/comments", taskProxy)
		protected.GET("/projects/:project_id/stats", taskProxy)
		protected.GET("/projects/:project_id/milestones", taskProxy)
		protected.POST("/projects/:project_id/milestones", taskProxy)
		protected.DELETE("/projects/:project_id/milestones/:milestone_id", taskProxy)
		protected.GET("/projects/:project_id/export/csv", taskProxy)
		protected.GET("/search", taskProxy)
	}

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

func reverseProxy(target string, cb *gobreaker.CircuitBreaker) gin.HandlerFunc {
	targetURL, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		rw.WriteHeader(http.StatusBadGateway) // 502 код
		rw.Write([]byte(fmt.Sprintf(`{"error": "Внутренний сервис %s недоступен"}`, targetURL.Host)))
	}

	return func(c *gin.Context) {
		_, err := cb.Execute(func() (interface{}, error) {

			proxy.ServeHTTP(c.Writer, c.Request)

			// сервис вернул 5xx - считаем его упавшим
			if c.Writer.Status() >= http.StatusInternalServerError {
				return nil, fmt.Errorf("микросервис вернул ошибку сервера")
			}
			return nil, nil
		})

		// цепь разомкнута - отдаём 503, не нагружая упавший сервис
		if err == gobreaker.ErrOpenState {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "Система перегружена. Включился предохранитель (Circuit Breaker). Подождите 7 секунд.",
				"state": "OPEN",
			})
		}
	}
}
