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

var tasks = []Task{}

var (
	lastID    int
	taskMutex sync.Mutex
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	err := LoadTasksFromFile("tasks.json")
	if err != nil {
		log.Fatalf("Failed to load tasks from tasks.json: %v", err)
	}
	logInfo("Starting server on http://localhost:8000")

	http.Handle("/tasks", LogRequestDuration(ValidateJSON(http.HandlerFunc(Tasks), http.MethodPost, http.MethodPut)))
	http.Handle("/tasks/", LogRequestDuration(ValidateJSON(http.HandlerFunc(Tasks), http.MethodPost, http.MethodPut)))
	http.Handle("/long/", LogRequestDuration(http.HandlerFunc(longRunningHandler)))
	http.HandleFunc("/tasks/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	doneChan := make(chan struct{})
	srv := &http.Server{
		Addr:    "0.0.0.0:8000",
		Handler: http.DefaultServeMux,
	}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		<-sigChan
		logInfo("Received shutdown signal, shutting down gracefully...")

		// Create a timeout context for the shutdown process
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Save tasks before shutdown
		if err := SaveTasksToFile("tasks.json"); err != nil {
			logError("Failed to save tasks to tasks.json: %v", err)
		} else {
			logInfo("Tasks saved to tasks.json")
		}

		// Attempt graceful shutdown
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}
		close(doneChan)
	}()

	if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
			logError("Task not found with ID %d in PUT", ID)
			writeJsonError(w, http.StatusNotFound, fmt.Sprintf("No task found with ID %d", ID))
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
			logError("Task not found with ID %d in DELETE", ID)
			writeJsonError(w, http.StatusNotFound, fmt.Sprintf("No task found with ID %d", ID))
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

func LoadTasksFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&tasks); err != nil {
		return err
	}
	logInfo("Tasks loaded successfully from %s", filename)

	// caluclate lastID
	lastID = 0
	for _, task := range tasks {
		if task.ID > lastID {
			lastID = task.ID
		}
	}

	return nil
}

func SaveTasksToFile(filename string) error {
	// Create backup of old tasks.json
	backupFilename := filename + ".bak"
	if _, err := os.Stat(filename); err == nil { // Check if file exists
		if err := os.Rename(filename, backupFilename); err != nil {
			logError("Warning: Failed to create backup %s: %v", backupFilename, err)
		} else {
			logInfo("Backup created: %s", backupFilename)
		}
	}

	// Overwrite original file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write JSON to file
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err = encoder.Encode(tasks); err != nil {
		return err
	}

	logInfo("Tasks successfully saved to %s", filename)
	return nil
}
