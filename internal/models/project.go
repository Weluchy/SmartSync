package models

// Константы ролей для типобезопасности
const (
	RoleOwner  = "owner"  // Вес 100: Создатель, может всё, включая удаление проекта
	RoleAdmin  = "admin"  // Вес 80: Зам. владельца, может управлять участниками и графом
	RoleEditor = "editor" // Вес 40: Исполнитель+, может менять оценки задач и зависимости
	RoleViewer = "viewer" // Вес 10: Наблюдатель, только чтение
)

// Карта весов для сравнения прав
var RoleWeights = map[string]int{
	RoleViewer: 10,
	RoleEditor: 40,
	RoleAdmin:  80,
	RoleOwner:  100,
}

type Project struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	OwnerID int    `json:"owner_id"`
	Role    string `json:"role,omitempty"`
}

type ProjectMember struct {
	ProjectID int    `json:"project_id"`
	UserID    int    `json:"user_id"`
	Username  string `json:"username,omitempty"`
	Role      string `json:"role"`
}
