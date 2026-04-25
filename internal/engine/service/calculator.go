package service

import (
	"context"
	"fmt"
	"smartsync/internal/engine/models"
	"smartsync/internal/engine/repository"
)

type CalculatorService struct {
	repo *repository.Storage
}

func NewCalculatorService(repo *repository.Storage) *CalculatorService {
	return &CalculatorService{repo: repo}
}

// Теперь функция принимает userID!
func (s *CalculatorService) RecalculateAll(ctx context.Context, userID int) error {
	tasks, _ := s.repo.GetAllTasks(userID)
	edges, _ := s.repo.GetAllEdges(userID)

	for _, t := range tasks {
		t.DurationHours = float64(t.Opt+4*t.Real+t.Pess) / 6.0
	}

	graphMap := make(map[int][]int)
	for _, e := range edges {
		graphMap[e.From] = append(graphMap[e.From], e.To)
	}

	memo := make(map[int]float64)
	visiting := make(map[int]bool)

	var dfs func(node int) (float64, error)
	dfs = func(node int) (float64, error) {
		if visiting[node] {
			return 0, fmt.Errorf("обнаружен цикл на задаче %d", node)
		}
		if val, exists := memo[node]; exists {
			return val, nil
		}

		visiting[node] = true
		maxChildPath := 0.0
		for _, child := range graphMap[node] {
			childWeight, err := dfs(child)
			if err != nil {
				return 0, err
			}
			if childWeight > maxChildPath {
				maxChildPath = childWeight
			}
		}
		visiting[node] = false

		weight := tasks[node].DurationHours + maxChildPath
		memo[node] = weight
		return weight, nil
	}

	var finalNodes []models.Task
	for id, t := range tasks {
		score, err := dfs(id)
		if err != nil {
			fmt.Printf("[ОШИБКА МАТЕМАТИКИ] %v. Пользователь ID: %d\n", err, userID)
			return err
		}
		t.PriorityScore = score
		finalNodes = append(finalNodes, *t)
		s.repo.UpdateTaskScores(t)
	}

	snapshot := models.GraphData{Nodes: finalNodes, Edges: edges}
	s.repo.CacheGraphSnapshot(ctx, userID, snapshot)

	fmt.Printf("[Math Engine] Граф пересчитан и закэширован для UserID: %d\n", userID)
	return nil
}
