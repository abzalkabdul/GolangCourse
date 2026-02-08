package main

import (
	"log"
	"net/http"
	"task-api/internal/handlers"
	"task-api/internal/middleware"
	"task-api/internal/models"
)

func main() {
	store := models.NewTaskStore()
	taskHandler := handlers.NewTaskHandler(store)

	mux := http.NewServeMux()

	mux.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			taskHandler.GetTask(w, r)
		case http.MethodPost:
			taskHandler.CreateTask(w, r)
		case http.MethodPatch:
			taskHandler.UpdateTask(w, r)
		case http.MethodDelete:
			taskHandler.DeleteTask(w, r)
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	handler := middleware.Logger(middleware.APIKeyAuth(mux))

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
