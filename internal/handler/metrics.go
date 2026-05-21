package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Счётчик запросов по методам и статусам
	requestCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "smartsync_requests_total",
			Help: "Общее количество запросов к Task Service",
		},
		[]string{"method", "endpoint", "status"},
	)

	// Гистограмма времени ответа (секунды)
	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "smartsync_request_duration_seconds",
			Help:    "Время выполнения запросов в секундах",
			Buckets: prometheus.DefBuckets, // 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10
		},
		[]string{"method", "endpoint"},
	)

	// Счётчик ошибок по типу
	errorCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "smartsync_errors_total",
			Help: "Количество ошибок по коду ответа",
		},
		[]string{"method", "endpoint", "status"},
	)

	// Счётчик активных WebSocket соединений
	activeWS = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "smartsync_websocket_active",
			Help: "Количество активных WebSocket соединений",
		},
	)
)

// PrometheusMiddleware собирает метрики по каждому запросу
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Пропускаем /metrics чтобы не зациклиться
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		c.Next()

		// Собираем данные после ответа
		latency := time.Since(start).Seconds()
		method := c.Request.Method
		endpoint := c.FullPath()
		status := strconv.Itoa(c.Writer.Status())

		// Инкрементим счётчики
		requestCounter.WithLabelValues(method, endpoint, status).Inc()
		requestDuration.WithLabelValues(method, endpoint).Observe(latency)

		if c.Writer.Status() >= 400 {
			errorCounter.WithLabelValues(method, endpoint, status).Inc()
		}
	}
}

// GetActiveWS возвращает счётчик активных WebSocket соединений
func GetActiveWS() prometheus.Gauge {
	return activeWS
}
