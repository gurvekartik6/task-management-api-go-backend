package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// Task Model
type Task struct {
	TaskId          int    `json:"task_id"`
	TaskName        string `json:"task_name"`
	TaskStatus      string `json:"task_status"`
	TaskDescription string `json:"task_description"`
}

// stroing tasks in slice
var tasks []Task

var nextId = 1

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Home Page")
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {

	sendJson(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

func sendJson(w http.ResponseWriter, statuscode int, data any) {

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(statuscode)

	err := json.NewEncoder(w).Encode(data)

	if err != nil {
		fmt.Fprintln(w, "Error Ocurred : ", err)
	}

}

// to handle all tasks
func TasksHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		sendJson(w, http.StatusOK, tasks)
	case http.MethodPost:

		var newTask Task

		err := json.NewDecoder(r.Body).Decode(&newTask)

		if err != nil {
			sendJson(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid json request body",
			})
			return
		}

		newTask.TaskName = strings.TrimSpace(newTask.TaskName)

		if newTask.TaskName == "" {
			sendJson(w, http.StatusBadRequest, map[string]string{
				"error": "Task Name Required",
			})
			return
		}

		newTask.TaskId = nextId
		nextId++

		if newTask.TaskStatus == "" {
			newTask.TaskStatus = "pending"
		}

		tasks = append(tasks, newTask)

		sendJson(w, http.StatusOK, newTask)

	default:
		w.Header().Set("Allow", "GET , POST")

		sendJson(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "Method Not Allowed",
		})
	}
}

func TaskHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "Get")

		sendJson(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})

		return
	}

	idString := strings.TrimPrefix(r.URL.Path, "/api/tasks/")

	id, err := strconv.Atoi(idString)

	if err != nil {
		sendJson(w, http.StatusBadRequest, map[string]string{
			"error": "ID required",
		})
		return
	}

	for _, task := range tasks {

		if task.TaskId == id {
			sendJson(w, http.StatusOK, task)
			return
		}
	}

	sendJson(w, http.StatusFound, map[string]string{
		"error": "task not found",
	})
}
func main() {

	fmt.Println("This is the main function ")

	http.HandleFunc("/", HomeHandler)
	http.HandleFunc("/health", HealthHandler)
	http.HandleFunc("/api/tasks", TasksHandler)
	http.HandleFunc("/api/tasks/", TaskHandler)

	fmt.Println("Server running on the http://localhost:8080")

	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		fmt.Println("Server Failed : ", err)
	}

}
