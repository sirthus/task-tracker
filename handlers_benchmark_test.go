package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func BenchmarkGetTasks(b *testing.B) {
	// Prepare initial tasks
	tasks = []Task{
		{ID: 1, Title: "Task 1", Completed: false},
		{ID: 2, Title: "Task 2", Completed: true},
	}

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rec := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Tasks(rec, req)
	}
}

func BenchmarkPostTasks(b *testing.B) {
	payload := `{"title":"Benchmark Task","completed":false}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		Tasks(rec, req)
	}
}

func BenchmarkPutTasks(b *testing.B) {
	// Prepare initial task
	tasks = []Task{
		{ID: 1, Title: "Initial Task", Completed: false},
	}

	payload := `{"title":"Updated Task","completed":true}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPut, "/tasks/1", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		Tasks(rec, req)
	}
}

func BenchmarkDeleteTasks(b *testing.B) {
	payload := Task{ID: 1, Title: "Task to Delete", Completed: false}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset tasks before each delete request
		tasks = []Task{payload}

		req := httptest.NewRequest(http.MethodDelete, "/tasks/1", nil)
		rec := httptest.NewRecorder()

		Tasks(rec, req)
	}
}
