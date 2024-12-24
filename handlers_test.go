package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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

var tests = []postTaskTestCase{
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

func TestGetTasks(t *testing.T) {
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
	for _, tt := range tests {
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
