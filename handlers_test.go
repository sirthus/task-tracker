package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

type invalidURLTestCase struct {
	name       string // Test case name
	method     string // HTTP method
	url        string // Endpoint
	wantStatus int    // Expected status code
}

type unsupportedMethodTestCase struct {
	name       string // Test case name
	method     string // HTTP method
	url        string // Endpoint
	wantStatus int    // Expected HTTP status code
}

type getTasksTestCase struct {
	name       string // test case name
	wantStatus int    // expected HTTP status code
	wantBody   string // expected response body
}

type postTaskTestCase struct {
	name       string // test case name
	payload    string // the json payload sent in request
	wantStatus int    // expected http status code
	wantBody   string // expected response body
}

type putTaskTestCase struct {
	name       string // Test case name
	id         string // Task ID to update
	payload    string // The JSON payload sent in the request
	wantStatus int    // Expected HTTP status code
	wantBody   string // Expected response body
}

type deleteTaskTestCase struct {
	name       string // Test case name
	id         string // Task ID to delete
	wantStatus int    // Expected HTTP status code
	wantBody   string // Expected response body
}

var invalidURLTests = []invalidURLTestCase{
	{
		name:       "Invalid Endpoint",
		method:     http.MethodGet,
		url:        "/invalid",
		wantStatus: http.StatusNotFound,
	},
	{
		name:       "Invalid Tasks Subpath",
		method:     http.MethodPost,
		url:        "/tasks/invalid",
		wantStatus: http.StatusNotFound,
	},
	{
		name:       "Empty Path",
		method:     http.MethodGet,
		url:        "/",
		wantStatus: http.StatusNotFound,
	},
}

var unsupportedMethodTests = []unsupportedMethodTestCase{
	{
		name:       "PUT on /tasks",
		method:     http.MethodPut,
		url:        "/tasks",
		wantStatus: http.StatusNotFound, // 404
	},
	{
		name:       "DELETE on /tasks",
		method:     http.MethodDelete,
		url:        "/tasks",
		wantStatus: http.StatusNotFound, // 404
	},
	{
		name:       "PATCH on /tasks/{id}",
		method:     http.MethodPatch,
		url:        "/tasks/1",
		wantStatus: http.StatusNotFound, // 404
	},
	{
		name:       "PATCH on /tasks/",
		method:     http.MethodPatch,
		url:        "/tasks/",
		wantStatus: http.StatusNotFound, // 404
	},
	{
		name:       "HEAD on /tasks/",
		method:     http.MethodHead,
		url:        "/tasks/",
		wantStatus: http.StatusNotFound, // 404 for unregistered method
	},
	{
		name:       "OPTIONS on /tasks/",
		method:     http.MethodOptions,
		url:        "/tasks/",
		wantStatus: http.StatusNotFound, // 404 for unregistered method
	},
	{
		name:       "Invalid HTTP Method on /tasks/",
		method:     "FOO",
		url:        "/tasks/",
		wantStatus: http.StatusNotFound, // 404 for invalid method
	},
}

var getTests = []getTasksTestCase{
	{
		name:       "Retrieve Tasks",
		wantStatus: http.StatusOK,
		wantBody:   `[{"id":1,"title":"Clean the carpet","completed":false},{"id":2,"title":"Pick up the groceries","completed":false},{"id":123,"title":"Doctor's appointment","completed":true}]`,
	},
	{
		name:       "No Tasks Available",
		wantStatus: http.StatusOK,
		wantBody:   `[]`,
	},
}

var postTests = []postTaskTestCase{
	{
		name:       "First Valid Task",
		payload:    `{"title": "New Task 1", "completed": false}`,
		wantStatus: http.StatusCreated,
		wantBody:   `{"id":124,"title":"New Task 1","completed":false}`,
	},
	{
		name:       "Second Valid Task",
		payload:    `{"title": "New Task 2", "completed": false}`,
		wantStatus: http.StatusCreated,
		wantBody:   `{"id":125,"title":"New Task 2","completed":false}`,
	},
	{
		name:       "Third Valid Task",
		payload:    `{"title": "New Task 3", "completed": false}`,
		wantStatus: http.StatusCreated,
		wantBody:   `{"id":126,"title":"New Task 3","completed":false}`,
	},
	{
		name:       "Task Without status",
		payload:    `{"title": "Task without status"}`,
		wantStatus: http.StatusCreated,
		wantBody:   `{"id":127,"title":"Task without status","completed":false}`,
	},
	{
		name:       "Task without title",
		payload:    `{"title": "", "completed": false}`,
		wantStatus: http.StatusBadRequest,
		wantBody:   `{"error":"Task title cannot be empty"}`,
	},
	{
		name:       "Empty task",
		payload:    `{}`,
		wantStatus: http.StatusBadRequest,
		wantBody:   `{"error":"Task title cannot be empty"}`,
	},
}

var putTests = []putTaskTestCase{
	{
		name:       "Update Existing Task",
		id:         "1",
		payload:    `{"title": "Updated Task", "completed": true}`,
		wantStatus: http.StatusOK,
		wantBody:   `{"id":1,"title":"Updated Task","completed":true}`,
	},
	{
		name:       "Task Not Found",
		id:         "999",
		payload:    `{"title": "Nonexistent Task", "completed": false}`,
		wantStatus: http.StatusNotFound,
		wantBody:   `{"error":"No task found with ID 999"}`,
	},
	{
		name:       "Invalid JSON",
		id:         "1",
		payload:    `{"title": "Missing Comma"`,
		wantStatus: http.StatusBadRequest,
		wantBody:   `{"error":"Invalid JSON format"}`,
	},
	{
		name:       "Empty Title",
		id:         "1",
		payload:    `{"title": "", "completed": false}`,
		wantStatus: http.StatusBadRequest,
		wantBody:   `{"error":"Task title cannot be empty"}`,
	},
	{
		name:       "Invalid ID",
		id:         "abc",
		payload:    `{"title": "Invalid ID", "completed": true}`,
		wantStatus: http.StatusBadRequest,
		wantBody:   `{"error":"Invalid Task ID"}`,
	},
	{
		name:       "Missing Status (Defaults to False)",
		id:         "1",
		payload:    `{"title": "Task Without Status"}`,
		wantStatus: http.StatusOK,
		wantBody:   `{"id":1,"title":"Task Without Status","completed":false}`,
	},
}

var deleteTests = []deleteTaskTestCase{
	{
		name:       "Delete Existing Task",
		id:         "1",
		wantStatus: http.StatusOK,
		wantBody:   `{"message":"Task deleted","status":"success"}`,
	},
	{
		name:       "Task Not Found",
		id:         "999",
		wantStatus: http.StatusNotFound,
		wantBody:   `{"error":"No task found with ID 999"}`,
	},
	{
		name:       "Invalid ID",
		id:         "abc",
		wantStatus: http.StatusBadRequest,
		wantBody:   `{"error":"Invalid Task ID"}`,
	},
}

func TestInvalidURLs(t *testing.T) {
	for _, tt := range invalidURLTests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request for the invalid URL
			req := httptest.NewRequest(tt.method, tt.url, nil)
			rec := httptest.NewRecorder()

			// Use the default handler to simulate the server behavior
			http.DefaultServeMux.ServeHTTP(rec, req)

			// Validate the status code
			if rec.Code != tt.wantStatus {
				t.Errorf("Test %s: got status %d, want %d", tt.name, rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestUnsupportedMethods(t *testing.T) {
	for _, tt := range unsupportedMethodTests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request with the unsupported method
			req := httptest.NewRequest(tt.method, tt.url, nil)
			rec := httptest.NewRecorder()

			// Call the handler
			http.DefaultServeMux.ServeHTTP(rec, req)

			// Validate the status code
			if rec.Code != tt.wantStatus {
				t.Errorf("Test %s: got status %d, want %d", tt.name, rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestGetTasks(t *testing.T) {
	tasks = []Task{
		{ID: 1, Title: "Clean the carpet", Completed: false},
		{ID: 2, Title: "Pick up the groceries", Completed: false},
		{ID: 123, Title: "Doctor's appointment", Completed: true},
	}

	for _, tt := range getTests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "No Tasks Available" {
				// Backup the original tasks slice
				originalTasks := tasks
				defer func() {
					tasks = originalTasks // Restore tasks after the test
				}()

				// Simulate no tasks
				tasks = []Task{}
			}

			// Simulate GET request
			req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
			rec := httptest.NewRecorder()

			// Call the handler
			Tasks(rec, req)

			// Validate the status code
			if rec.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", rec.Code, tt.wantStatus)
			}

			// Validate the response body
			gotBody := strings.TrimSpace(rec.Body.String())
			if gotBody != tt.wantBody {
				t.Errorf("got body %s, want %s", gotBody, tt.wantBody)
			}
		})
	}
}

func TestCreateTask(t *testing.T) {
	lastID = 123 // Initialize lastID correctly
	tasks = []Task{}

	for _, tt := range postTests {
		t.Run(tt.name, func(t *testing.T) {
			// create request
			req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(tt.payload))

			// capture the response
			rec := httptest.NewRecorder()

			// call the handler
			Tasks(rec, req)

			// validate the status code
			if rec.Code != tt.wantStatus {
				t.Errorf("Test %s: got status %d, want %d; payload: %s", tt.name, rec.Code, tt.wantStatus, tt.payload)
			}

			// validate response body
			gotBody := strings.TrimSpace(rec.Body.String())
			if gotBody != tt.wantBody {
				t.Errorf("Test %s: got body %s, want %s; payload: %s", tt.name, gotBody, tt.wantBody, tt.payload)
			}

		})
	}
}

func TestUpdateTask(t *testing.T) {
	tasks = []Task{
		{ID: 1, Title: "Clean the carpet", Completed: false},
		{ID: 2, Title: "Pick up the groceries", Completed: false},
		{ID: 123, Title: "Doctor's appointment", Completed: true},
	}

	for _, tt := range putTests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the request
			req := httptest.NewRequest(http.MethodPut, "/tasks/"+tt.id, strings.NewReader(tt.payload))
			rec := httptest.NewRecorder()

			// Call the handler
			Tasks(rec, req)

			// Validate the status code
			if rec.Code != tt.wantStatus {
				t.Errorf("Test %s: got status %d, want %d", tt.name, rec.Code, tt.wantStatus)
			}

			// Validate the response body
			gotBody := strings.TrimSpace(rec.Body.String())
			if gotBody != tt.wantBody {
				t.Errorf("Test %s: got body %s, want %s", tt.name, gotBody, tt.wantBody)
			}
		})
	}
}

func TestDeleteTask(t *testing.T) {
	tasks = []Task{
		{ID: 1, Title: "Clean the carpet", Completed: false},
		{ID: 2, Title: "Pick up the groceries", Completed: false},
		{ID: 123, Title: "Doctor's appointment", Completed: true},
	}

	for _, tt := range deleteTests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the request
			req := httptest.NewRequest(http.MethodDelete, "/tasks/"+tt.id, nil)
			rec := httptest.NewRecorder()

			// Call the handler
			Tasks(rec, req)

			// Validate the status code
			if rec.Code != tt.wantStatus {
				t.Errorf("Test %s: got status %d, want %d", tt.name, rec.Code, tt.wantStatus)
			}

			// Validate the response body
			gotBody := strings.TrimSpace(rec.Body.String())
			if gotBody != tt.wantBody {
				t.Errorf("Test %s: got body %s, want %s", tt.name, gotBody, tt.wantBody)
			}
		})
	}
}

// func TestTasksConcurrency(t *testing.T) {
// 	// Start with an empty tasks slice
// 	tasks = []Task{}
// 	lastID = 123 // Start IDs from 124

// 	var wg sync.WaitGroup
// 	const numGoroutines = 100

// 	// Simulate concurrent POST requests
// 	for i := 0; i < numGoroutines; i++ {
// 		wg.Add(1)
// 		go func(id int) {
// 			defer wg.Done()

// 			payload := fmt.Sprintf(`{"title": "Task %d", "completed": false}`, id)
// 			req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(payload))
// 			rec := httptest.NewRecorder()

// 			Tasks(rec, req)

// 			// Validate the status code
// 			if rec.Code != http.StatusCreated {
// 				t.Errorf("POST failed for goroutine %d: got status %d, want %d", id, rec.Code, http.StatusCreated)
// 			}
// 		}(i)
// 	}

// 	// Wait for all POST requests to complete
// 	wg.Wait()

// 	// Validate the number of tasks
// 	if len(tasks) != numGoroutines {
// 		t.Errorf("Expected %d tasks, got %d", numGoroutines, len(tasks))
// 	}

// 	// Validate sequential IDs (lastID currently hardcoded to 123)
// 	startingID := 124
// 	for i, task := range tasks {
// 		expectedID := startingID + i
// 		if task.ID != expectedID {
// 			t.Errorf("Task ID mismatch at index %d: got %d, want %d", i, task.ID, expectedID)
// 		}
// 	}
// }

func TestLoadAndSaveTasks(t *testing.T) {
	tempFile := "test_tasks.json"
	defer os.Remove(tempFile)

	// Test saving tasks
	tasks = []Task{
		{ID: 1, Title: "Task 1", Completed: false},
		{ID: 2, Title: "Task 2", Completed: true},
	}
	if err := SaveTasksToFile(tempFile); err != nil {
		t.Fatalf("Failed to save tasks: %v", err)
	}

	// Clear the current tasks and test loading from the file
	tasks = nil
	if err := LoadTasksFromFile(tempFile); err != nil {
		t.Fatalf("Failed to load tasks: %v", err)
	}

	// Validate loaded tasks
	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(tasks))
	}
	if tasks[0].Title != "Task 1" || tasks[1].Completed != true {
		t.Errorf("Loaded tasks do not match expected values")
	}
}

func TestLoadTasksFromNonExistentFile(t *testing.T) {
	nonExistentFile := "nonexistent_tasks.json"

	err := LoadTasksFromFile(nonExistentFile)
	if err == nil {
		t.Errorf("Expected an error when loading from a non-existent file, got nil")
	}
}

func TestSaveTasksCreatesBackup(t *testing.T) {
	tempFile := "test_tasks.json"
	backupFile := tempFile + ".bak"
	defer os.Remove(tempFile)
	defer os.Remove(backupFile)

	// Initial save
	tasks = []Task{
		{ID: 1, Title: "Original Task", Completed: false},
	}
	if err := SaveTasksToFile(tempFile); err != nil {
		t.Fatalf("Failed to save tasks: %v", err)
	}

	// Modify tasks and save again
	tasks = []Task{
		{ID: 2, Title: "Updated Task", Completed: true},
	}
	if err := SaveTasksToFile(tempFile); err != nil {
		t.Fatalf("Failed to save tasks again: %v", err)
	}

	// Validate backup file
	backupData, err := os.ReadFile(backupFile)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}
	if !strings.Contains(string(backupData), "Original Task") {
		t.Errorf("Backup file does not contain original tasks")
	}
}

func TestSaveTasksToInvalidLocation(t *testing.T) {
	invalidFile := "/invalid_path/test_tasks.json"

	err := SaveTasksToFile(invalidFile)
	if err == nil {
		t.Errorf("Expected an error when saving to an invalid location, got nil")
	}
}
