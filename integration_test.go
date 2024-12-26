package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIntegrationWorkFlow(t *testing.T) {
	// Reset global state for testing
	tasks = []Task{}
	lastID = 0

	// Step 1: Test POST /tasks
	reqBody := bytes.NewBuffer([]byte(`{"title":"Test Task","completed":false}`))
	req := httptest.NewRequest(http.MethodPost, "/tasks", reqBody)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	Tasks(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	expected := `{"id":1,"title":"Test Task","completed":false}`
	actual := strings.TrimSpace(rec.Body.String())
	if actual != expected {
		t.Fatalf("expected body %s, got %s", expected, actual)
	}

	// Step 2: Test GET /tasks
	req = httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rec = httptest.NewRecorder()

	Tasks(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	expected = `[{"id":1,"title":"Test Task","completed":false}]`
	actual = strings.TrimSpace(rec.Body.String())
	if actual != expected {
		t.Fatalf("expected body %s, got %s", expected, actual)
	}

	// Step 3: Test PUT /tasks/1
	reqBody = bytes.NewBuffer([]byte(`{"title":"Updated Task","completed":true}`))
	req = httptest.NewRequest(http.MethodPut, "/tasks/1", reqBody)
	req.Header.Set("Content-Type", "application/json")

	rec = httptest.NewRecorder()
	Tasks(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	expected = `{"id":1,"title":"Updated Task","completed":true}`
	actual = strings.TrimSpace(rec.Body.String())
	if actual != expected {
		t.Fatalf("expected body %s, got %s", expected, actual)
	}

	// Step 4: Test DELETE /tasks/1
	req = httptest.NewRequest(http.MethodDelete, "/tasks/1", nil)
	rec = httptest.NewRecorder()

	Tasks(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	expected = `{"message":"Task deleted","status":"success"}`
	actual = strings.TrimSpace(rec.Body.String())
	if actual != expected {
		t.Fatalf("expected body %s, got %s", expected, actual)
	}

	// Step 5: Confirm task is deleted via GET /tasks
	req = httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rec = httptest.NewRecorder()

	Tasks(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	expected = `[]`
	actual = strings.TrimSpace(rec.Body.String())
	if actual != expected {
		t.Fatalf("expected body %s, got %s", expected, actual)
	}
}
