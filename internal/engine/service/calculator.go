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

// Главная функция пересчета (CQRS Command)
func (s *CalculatorService) RecalculateAll(ctx context.Context) error {
	tasks, _ := s.repo.GetAllTasks()
	edges, _ := s.repo.GetAllEdges()

	// 1. Вычисляем PERT (Ожидаемое время)
	for _, t := range tasks {
		t.DurationHours = float64(t.Opt+4*t.Real+t.Pess) / 6.0
	}

	// 2. Строим карту смежности для графа
	graphMap := make(map[int][]int)
	for _, e := range edges {
		graphMap[e.From] = append(graphMap[e.From], e.To)
	}

	// 3. Алгоритм DFS с защитой от циклов
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
			} // Проброс ошибки цикла
			if childWeight > maxChildPath {
				maxChildPath = childWeight
			}
		}
		visiting[node] = false

		weight := tasks[node].DurationHours + maxChildPath
		memo[node] = weight
		return weight, nil
	}

	// 4. Применяем расчеты ко всем узлам
	var finalNodes []models.Task
	for id, t := range tasks {
		score, err := dfs(id)
		if err != nil {
			fmt.Printf("[ОШИБКА МАТЕМАТИКИ] %v. Расчет прерван.\n", err)
			return err
		}
		t.PriorityScore = score
		finalNodes = append(finalNodes, *t)

		// Сохраняем новые цифры в БД
		s.repo.UpdateTaskScores(t)
	}

	// 5. Кэшируем готовый слепок для фронтенда
	snapshot := models.GraphData{Nodes: finalNodes, Edges: edges}
	s.repo.CacheGraphSnapshot(ctx, snapshot)

	fmt.Println("[Math Engine] Граф пересчитан и закэширован")
	return nil
}
