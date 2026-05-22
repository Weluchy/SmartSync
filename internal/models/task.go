package models

import "time"

type Task struct {
	ID            int       `json:"id"`
	ProjectID     int       `json:"project_id"`
	UserID        int       `json:"user_id"`
	AssigneeID    *int      `json:"assignee_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"` // <-- Вот эта строка была нужна!
	Status        string    `json:"status"`
	Opt           int       `json:"opt"`
	Real          int       `json:"real"`
	Pess          int       `json:"pess"`
	DurationHours float64   `json:"duration_hours"`
	PriorityScore float64   `json:"priority_score"`
	CreatedAt     time.Time `json:"created_at"`

	// Эти поля заполняются динамически через сервис
	CreatedByName string `json:"created_by_name"`
	AssigneeName  string `json:"assignee_name"`
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
