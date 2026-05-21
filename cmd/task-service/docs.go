// Package main Task Service API
//
// SmartSync — микросервис управления задачами.
// Отвечает за CRUD задач, зависимости (граф), права доступа (RBAC).
//
// Архитектура:
//   - Postgres: хранение задач, пользователей, проектов
//   - Redis: кэширование графа задач
//   - NATS: отправка событий аудита и обновлений
//   - Prometheus: мониторинг метрик (/metrics)
//
// TermsOfService: http://swagger.io/terms/
// Contact: Артем <student@bsu.by>
// Version: 1.0.0
// Host: localhost:8080
// BasePath: /
//
// SecurityDefinitions:
//
//	bearerAuth:
//	  type: apiKey
//	  name: Authorization
//	  in: header
//	  description: "Введите токен в формате: Bearer <token>"
//
// Schemes: http
package main
