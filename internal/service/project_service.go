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
