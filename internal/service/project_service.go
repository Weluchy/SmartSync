package service

import (
	"smartsync/internal/models"
	"smartsync/internal/repository"
)

type ProjectService struct {
	repo *repository.ProjectRepository
}

func NewProjectService(repo *repository.ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) GetUserProjects(userID int) ([]models.Project, error) {
	return s.repo.GetUserProjects(userID)
}

func (s *ProjectService) CreateProject(name string, userID int) (int, error) {
	return s.repo.CreateProject(name, userID)
}

func (s *ProjectService) DeleteProject(projectID, userID int) error {
	return s.repo.DeleteProject(projectID, userID)
}

func (s *ProjectService) RenameProject(projectID, userID int, newName string) error {
	return s.repo.RenameProject(projectID, userID, newName)
}

// ИСПРАВЛЕНО: Добавлен параметр role
func (s *ProjectService) AddMember(projectID, ownerID int, username string, role string) error {
	return s.repo.AddMember(projectID, ownerID, username, role)
}

func (s *ProjectService) GetProjectMembers(projectID int) ([]models.ProjectMember, error) {
	return s.repo.GetProjectMembers(projectID)
}

func (s *ProjectService) RemoveMember(projectID, ownerID, targetUserID int) error {
	return s.repo.RemoveMember(projectID, ownerID, targetUserID)
}

func (s *ProjectService) UpdateMemberRole(projectID, ownerID, targetUserID int, newRole string) error {
	return s.repo.UpdateMemberRole(projectID, ownerID, targetUserID, newRole)
}
