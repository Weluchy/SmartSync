# SmartSync - Event-Driven Task Management Platform

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-316192?style=flat&logo=postgresql&logoColor=white)
![MongoDB](https://img.shields.io/badge/MongoDB-4EA94B?style=flat&logo=mongodb&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-DC382D?style=flat&logo=redis&logoColor=white)
![NATS](https://img.shields.io/badge/NATS-27A1E1?style=flat&logo=nats&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=flat&logo=docker&logoColor=white)

**SmartSync** - it is a microservice platform for project and task management (Kanban). The project is designed with an emphasis on Event-Driven architecture, and the use of mathematical methods of task evaluation (PERT) to build a dependency graph and identify a critical path.

## ✨ Key architectural solutions

* **API Gateway Pattern:** A single entry point with implemented Rate Limiter (DDoS protection) and Circuit Breaker (gobreaker) pattern to prevent cascading failures.
* **Event-oriented architecture:** Microservices communicate asynchronously through the message broker **NATS**.
* **The math engine (Priority Service):** A depth-first graph traversal (DFS) algorithm for automatically recalculating priorities and deadlines for tasks.
* **Immutable audit (Immutable Audit):** A dedicated audit service based on **MongoDB**, which records all events in the system in isolation.
* **Real-time обновления:** Broadcasting events (creating connections, changing statuses) to clients via WebSockets directly from NATS.
* **Monitoring:** Integration with Prometheus and Grafana to collect performance metrics.

## 🏗 The structure of microservices

1. **`gateway`** — routing, JWT authorization, WebSocket connections, Rate Limiting and Circuit Breaker.
2. **`auth-service`** — user authentication and issuance of JWT tokens.
3. **`task-service`** is the main CRUD of tasks, projects, milestones, and dependency graph management (uses PostgreSQL and Redis).
4. **`audit—service`** is a NATS event viewer that stores the history of actions in MongoDB.
5. **`priority-service`** is a background calculator for the task graph and critical path (PERT).
6. **`deadline-service`** is a worker that tracks task deadlines and generates alerts.

## 🚀Quick start

The project is fully containerized. The entire infrastructure (backend, frontend, database, and broker) is raised by a single team.

```bash
# 1. Cloning a repository
git clone [https://github.com/твое-имя/SmartSync.git](https://github.com/твое-имя/SmartSync.git)
cd SmartSync

# 2. Launching the application and infrastructure via Docker Compose
docker-compose up -d --build
