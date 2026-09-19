package service

import (
	"smartsync/internal/engine/repository"
	"time"
)

type Calculator struct {
	repo *repository.Storage
}

func NewCalculator(repo *repository.Storage) *Calculator {
	return &Calculator{repo: repo}
}

func (c *Calculator) RecalculateGraph(projectID int) {
	tasks, err := c.repo.GetProjectTasks(projectID)
	if err != nil || len(tasks) == 0 {
		return
	}
	edges, _ := c.repo.GetTaskDependencies(projectID)

	adj := make(map[int][]int)
	inDegree := make(map[int]int)
	durations := make(map[int]float64)
	earlyFinish := make(map[int]float64)
	earlyStart := make(map[int]float64)
	createdAt := make(map[int]time.Time)

	// начальные значения по каждой задаче
	for _, t := range tasks {
		inDegree[t.ID] = 0
		if t.Status == "done" {
			durations[t.ID] = 0.0
		} else {
			durations[t.ID] = float64(t.Opt+4*t.Real+t.Pess) / 6.0
		}
		earlyStart[t.ID] = 0.0
		earlyFinish[t.ID] = durations[t.ID]
		createdAt[t.ID] = t.CreatedAt
	}

	// e.from - родитель, e.to - ребёнок
	for _, e := range edges {
		adj[e.From] = append(adj[e.From], e.To)
		inDegree[e.To]++
	}

	// топологическая сортировка (кан)
	var queue []int
	for _, t := range tasks {
		if inDegree[t.ID] == 0 {
			queue = append(queue, t.ID)
		}
	}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		for _, child := range adj[curr] {
			// родитель даёт более поздний старт - обновляем ребёнка
			startFromParent := earlyFinish[curr]
			if startFromParent > earlyStart[child] {
				earlyStart[child] = startFromParent
				earlyFinish[child] = startFromParent + durations[child]
			}
			inDegree[child]--
			if inDegree[child] == 0 {
				queue = append(queue, child)
			}
		}
	}

	// дедлайны считаем от самой ранней корневой задачи
	var baseTime time.Time
	for _, t := range tasks {
		if inDegree[t.ID] == 0 { // inDegree тут как флаг "корневая"
			if baseTime.IsZero() || createdAt[t.ID].Before(baseTime) {
				baseTime = createdAt[t.ID]
			}
		}
	}
	if baseTime.IsZero() {
		// нет корневых - берём самое раннее время
		for _, t := range tasks {
			if baseTime.IsZero() || createdAt[t.ID].Before(baseTime) {
				baseTime = createdAt[t.ID]
			}
		}
	}

	for _, t := range tasks {
		// earlyFinish в часах -> unix ms
		deadlineUnix := baseTime.UnixMilli() + int64(earlyFinish[t.ID]*3600000)
		c.repo.UpdateTaskMetrics(t.ID, durations[t.ID], earlyFinish[t.ID], deadlineUnix)
	}
}
