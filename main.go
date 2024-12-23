// Task Tracker http server allows tracking tasks and their completion statuses
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Task struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

var tasks = []Task{
	{ID: 1, Title: "Clean the carpet", Completed: false},
	{ID: 2, Title: "Pick up the groceries", Completed: false},
	{ID: 123, Title: "Doctor's appointment", Completed: true},
}

var (
	lastID    = 123
	taskMutex sync.Mutex
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	logInfo("Starting server on http://localhost:8000")

	http.HandleFunc("/tasks/", Tasks)
	http.HandleFunc("/long", longRunningHandler)
	doneChan := make(chan struct{})
	srv := &http.Server{
		Addr:    "localhost:8000",
		Handler: http.DefaultServeMux,
	}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down gracefully...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}
		close(doneChan)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Listen failed: %v", err)
	}

	<-doneChan // Wait for shutdown signal
	log.Println("Server shutdown complete.")
}

func longRunningHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting long-running request...")
	time.Sleep(10 * time.Second) // Simulate processing delay
	log.Println("Finished long-running request.")
	w.Write([]byte("Request completed"))
}

// Tasks handles requests to retrieve, create, or delete tasks via HTTP methods.
func Tasks(w http.ResponseWriter, r *http.Request) {
	// Prints log to Stdout
	logInfo("Received %s request for %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	// Check that method type is supported
	if r.Method != "GET" && r.Method != "POST" && r.Method != "PUT" && r.Method != "DELETE" {
		logError("Unsupported method: %s", r.Method)
		writeJsonError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}
	switch r.Method {
	case "GET":
		// Marshal tasks struct into valid json
		taskMutex.Lock()
		defer taskMutex.Unlock()
		jsonData, err := json.Marshal(tasks)
		if err != nil {
			logError("JSON marshalling failed")
			writeJsonError(w, http.StatusInternalServerError, "Internal server error: JSON marshalling failed")
			return
		}
		// Specify response format as JSON to ensure correct client parsing
		w.Header().Set("Content-Type", "application/json")
		// Writes the json data to the client
		w.Write(jsonData)
	case "POST":
		// Reads the body for valid json to add as new task
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logError("Failed to read request body")
			writeJsonError(w, http.StatusBadRequest, "Failed to read request body")
			return
		}

		var newTask Task
		// Unmarshals json into struct fields
		err = json.Unmarshal(body, &newTask)
		if err != nil {
			logError("Invalid JSON Format in POST request")
			writeJsonError(w, http.StatusBadRequest, "Invalid JSON format")
			return
		}
		if newTask.Title == "" {
			logError("Invalid task title in POST request")
			writeJsonError(w, http.StatusBadRequest, "Task title cannot be empty")
			return
		}
		// lastID tracks the ID of the most recently added task
		taskMutex.Lock()
		defer taskMutex.Unlock()
		lastID++
		newTask.ID = lastID
		// Add new task to tasks
		tasks = append(tasks, newTask)
		w.Header().Set("Content-Type", "application/json")
		// Sets status to 201 to acknowledge task creation
		w.WriteHeader(http.StatusCreated)
		// Writes new task back to client
		json.NewEncoder(w).Encode(newTask)
	case "PUT":
		ID, err := ParseTaskID(r)
		if err != nil {
			logError(err.Error())
			writeJsonError(w, http.StatusBadRequest, err.Error())
			return
		}
		// Reads the body for valid json to add as new task
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logError("Failed to read request body in PUT")
			writeJsonError(w, http.StatusBadRequest, "Failed to read request body")
			return
		}
		var newTask Task
		// Unmarshals json into struct fields
		err = json.Unmarshal(body, &newTask)
		if err != nil {
			logError("Invalid JSON format in PUT")
			writeJsonError(w, http.StatusBadRequest, "Invalid JSON format")
			return
		}
		if newTask.Title == "" {
			logError("Empty task title in PUT")
			writeJsonError(w, http.StatusBadRequest, "Task title cannot be empty")
			return
		}
		// taskFound tracks whether specified task number exists
		taskFound := false
		// Removes specified task if found
		taskMutex.Lock()
		defer taskMutex.Unlock()
		var updatedIndex int
		for i, t := range tasks {
			if t.ID == ID {
				tasks[i].Title = newTask.Title
				tasks[i].Completed = newTask.Completed
				updatedIndex = i
				taskFound = true
				break
			}
		}
		if !taskFound {
			logError("Task not found in PUT")
			writeJsonError(w, http.StatusNotFound, "Task not found")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		// Outputs success message in json format
		json.NewEncoder(w).Encode(tasks[updatedIndex])

	case "DELETE":
		ID, err := ParseTaskID(r)
		if err != nil {
			logError(err.Error())
			writeJsonError(w, http.StatusBadRequest, err.Error())
			return
		}
		// taskFound tracks whether specified task number exists
		taskFound := false
		// Removes specified task if found
		taskMutex.Lock()
		defer taskMutex.Unlock()
		for i, t := range tasks {
			if t.ID == ID {
				tasks = append(tasks[:i], tasks[i+1:]...)
				taskFound = true
				break
			}
		}
		if !taskFound {
			logError("Task not found in DELETE")
			writeJsonError(w, http.StatusNotFound, "Task not found")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		// Outputs success message in json format
		json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Task deleted"})
	}
}

func writeJsonError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func logInfo(msg string, args ...interface{}) {
	log.Printf("[INFO] "+msg, args...)
}

func logError(msg string, args ...interface{}) {
	log.Printf("[Error] "+msg, args...)
}

func ParseTaskID(r *http.Request) (int, error) {
	// Cleans path to allow trailing slashes
	r.URL.Path = path.Clean(r.URL.Path)
	// Splits URL based on /
	parts := strings.Split(r.URL.Path, "/")
	// Checks that URL is properly formed "host/tasks/{id}"
	if len(parts) != 3 {
		return 0, fmt.Errorf("Invalid URL")
	}
	// Converts task number to integer
	ID, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return 0, fmt.Errorf("Invalid Task ID")
	}
	return ID, nil
}
