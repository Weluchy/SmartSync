package models

import "time"

type Milestone struct {
	ID          int       `json:"id"`
	ProjectID   int       `json:"project_id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Deadline    time.Time `json:"deadline"`
	CreatedAt   time.Time `json:"created_at"`
}
