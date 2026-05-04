package models

import "time"

type Task struct {
	ID            int       `json:"id"`
	Title         string    `json:"title"`
	Opt           int       `json:"opt"`
	Real          int       `json:"real"`
	Pess          int       `json:"pess"`
	DurationHours float64   `json:"duration_hours"`
	PriorityScore float64   `json:"priority_score"`
	UserID        int       `json:"user_id"`         // Это и есть ID создателя
	CreatedByName string    `json:"created_by_name"` // ДОБАВЛЯЕМ ТОЛЬКО ЭТО ДЛЯ ФРОНТЕНДА
	ProjectID     int       `json:"project_id"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"` // Для сортировки по времени
}
type GraphEdge struct {
	From int `json:"from"`
	To   int `json:"to"`
}

type GraphData struct {
	Nodes []Task      `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}
type Dependency struct {
	TaskID      int `json:"task_id"`
	DependsOnID int `json:"depends_on_id"`
}
