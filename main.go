package main

import (
	"encoding/json"
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

func main() {
	http.HandleFunc("/tasks", Tasks)
	http.ListenAndServe("localhost:8000", nil)
}

func Tasks(w http.ResponseWriter, r *http.Request) {

	log.Printf("%s %s", r.Method, r.URL.Path)

	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	jsonData, err := json.Marshal(tasks)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	w.Write(jsonData)
}
