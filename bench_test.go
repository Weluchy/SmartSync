package main

import (
	"testing"
)

// Имитация алгоритма DFS для замера производительности
func runDFS(nodeCount int) {
	graph := make(map[int][]int)
	durations := make(map[int]int)

	// Генерируем тестовый граф
	for i := 0; i < nodeCount; i++ {
		durations[i] = 10
		if i > 0 {
			graph[i-1] = append(graph[i-1], i)
		}
	}

	memo := make(map[int]float64)
	var dfs func(int) float64
	dfs = func(node int) float64 {
		if val, exists := memo[node]; exists {
			return val
		}
		maxChild := 0.0
		for _, child := range graph[node] {
			childWeight := dfs(child)
			if childWeight > maxChild {
				maxChild = childWeight
			}
		}
		weight := float64(durations[node]) + maxChild
		memo[node] = weight
		return weight
	}

	for i := 0; i < nodeCount; i++ {
		dfs(i)
	}
}

// Бенчмарк для 100 задач
func BenchmarkDFS_100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runDFS(100)
	}
}

// Бенчмарк для 1000 задач
func BenchmarkDFS_1000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runDFS(1000)
	}
}

// Бенчмарк для 10000 задач
func BenchmarkDFS_10000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runDFS(10000)
	}
}
