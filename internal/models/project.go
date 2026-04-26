package models

type Project struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	OwnerID int    `json:"owner_id"`
	Role    string `json:"role,omitempty"` // Какая роль у того, кто сейчас смотрит проект
}

// Участник проекта (для совместного доступа)
type ProjectMember struct {
	ProjectID int    `json:"project_id"`
	UserID    int    `json:"user_id"`
	Username  string `json:"username,omitempty"`
	Role      string `json:"role"` // Роли: "owner", "editor", "viewer"
}
