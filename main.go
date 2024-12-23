package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
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

var lastID = 123

func main() {
	http.HandleFunc("/tasks", Tasks)
	http.ListenAndServe("localhost:8000", nil)
}

func Tasks(w http.ResponseWriter, r *http.Request) {

	log.Printf("%s %s", r.Method, r.URL.Path)

	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	switch r.Method {
	case "GET":
		jsonData, err := json.Marshal(tasks)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")

		w.Write(jsonData)
	case "POST":
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading data.", http.StatusBadRequest)
			return
		}

		var newTask Task
		err = json.Unmarshal(body, &newTask)
		if err != nil {
			http.Error(w, "Error reading data.", http.StatusBadRequest)
			return
		}
		lastID++
		newTask.ID = lastID

		tasks = append(tasks, newTask)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newTask)
	}

}
