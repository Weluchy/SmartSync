package models

type Task struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	Opt           int     `json:"opt"`
	Real          int     `json:"real"`
	Pess          int     `json:"pess"`
	DurationHours float64 `json:"duration_hours"`
	PriorityScore float64 `json:"priority_score"`
	UserID        int     `json:"user_id"`
	ProjectID     int     `json:"project_id"`
	Status        string  `json:"status"`
}

type GraphEdge struct {
	From int `json:"from"`
	To   int `json:"to"`
}

type GraphData struct {
	Nodes []Task      `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}
