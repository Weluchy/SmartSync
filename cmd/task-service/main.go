package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
)

type Task struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	Opt           int     `json:"opt"`
	Real          int     `json:"real"`
	Pess          int     `json:"pess"`
	DurationHours float64 `json:"duration_hours"`
	PriorityScore float64 `json:"priority_score"`
}

type Dependency struct {
	DependsOnID int `json:"depends_on_id"`
}

type GraphEdge struct {
	From int `json:"from"`
	To   int `json:"to"`
}

type GraphData struct {
	Nodes []Task      `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

func main() {
	connStr := "postgres://user:password@127.0.0.1:5433/smartsync?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	r := gin.Default()

	// Настройка CORS (добавили разрешение на DELETE-запросы)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 1. Создание задачи
	r.POST("/tasks", func(c *gin.Context) {
		var t Task
		if err := c.ShouldBindJSON(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if t.Opt == 0 {
			t.Opt = 1
		}
		if t.Real == 0 {
			t.Real = 1
		}
		if t.Pess == 0 {
			t.Pess = 1
		}

		var id int
		err := db.QueryRow("INSERT INTO tasks (title, opt, real, pess) VALUES ($1, $2, $3, $4) RETURNING id",
			t.Title, t.Opt, t.Real, t.Pess).Scan(&id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		nc.Publish("task.created", []byte(fmt.Sprintf("%d", id)))
		c.JSON(http.StatusOK, gin.H{"message": "Задача создана", "id": id})
	})

	// 2. Создание связи
	r.POST("/tasks/:id/dependencies", func(c *gin.Context) {
		taskID, _ := strconv.Atoi(c.Param("id"))
		var dep Dependency
		c.ShouldBindJSON(&dep)
		db.Exec("INSERT INTO dependencies (task_id, depends_on_id) VALUES ($1, $2)", taskID, dep.DependsOnID)
		nc.Publish("graph.updated", []byte("updated"))
		c.JSON(http.StatusOK, gin.H{"message": "Связь создана"})
	})

	// 3. Очистка графа (Сброс связей)
	r.DELETE("/dependencies", func(c *gin.Context) {
		_, err := db.Exec("TRUNCATE dependencies")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении связей"})
			return
		}
		nc.Publish("graph.updated", []byte("reset"))
		c.JSON(http.StatusOK, gin.H{"message": "Граф сброшен"})
	})

	// 4. Отдача Графа
	r.GET("/graph", func(c *gin.Context) {
		var graph GraphData

		// Теперь мы достаем opt, real, pess из базы, чтобы фронт мог считать риски (сигму)!
		rowsNodes, _ := db.Query("SELECT id, title, opt, real, pess, duration_hours, priority_score FROM tasks")
		for rowsNodes.Next() {
			var t Task
			rowsNodes.Scan(&t.ID, &t.Title, &t.Opt, &t.Real, &t.Pess, &t.DurationHours, &t.PriorityScore)
			graph.Nodes = append(graph.Nodes, t)
		}
		rowsNodes.Close()

		rowsEdges, _ := db.Query("SELECT depends_on_id, task_id FROM dependencies")
		for rowsEdges.Next() {
			var e GraphEdge
			rowsEdges.Scan(&e.From, &e.To)
			graph.Edges = append(graph.Edges, e)
		}
		rowsEdges.Close()

		c.JSON(http.StatusOK, graph)
	})

	fmt.Println("Task Service [PERT + Delete Edition] запущен")
	r.Run(":8080")
}
