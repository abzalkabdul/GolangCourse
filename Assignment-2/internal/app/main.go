package app

import (
	"assignment-2/internal/handler"
	"assignment-2/internal/middleware"
	"assignment-2/internal/repository"
	"assignment-2/internal/repository/_postgres"
	"assignment-2/internal/usecase"
	"assignment-2/pkg/modules"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbConfig := initPostgreConfig()

	pg := _postgres.NewPGXDialect(ctx, dbConfig)
	repos := repository.NewRepositories(pg)

	// Original debug: print all users on startup
	users, err := repos.GetUsers()
	if err != nil {
		fmt.Printf("Error fetching users: %v\n", err)
	} else {
		fmt.Printf("Users: %+v\n", users)
	}

	userUC := usecase.NewUserUsecase(repos.UserRepository)
	userHandler := handler.NewUserHandler(userUC)

	mux := http.NewServeMux()

	// Health check – no auth required
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// User routes – all protected by Auth middleware
	mux.Handle("/users", middleware.Auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userHandler.GetUsers(w, r)
		case http.MethodPost:
			userHandler.CreateUser(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/users/", middleware.Auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Reject if path is just /users/ with no ID
		trimmed := strings.TrimPrefix(r.URL.Path, "/users/")
		if trimmed == "" {
			http.NotFound(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			userHandler.GetUserByID(w, r)
		case http.MethodPatch, http.MethodPut:
			userHandler.UpdateUser(w, r)
		case http.MethodDelete:
			userHandler.DeleteUser(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Wrap entire mux with the Logger middleware
	loggedMux := middleware.Logger(mux)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", loggedMux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func initPostgreConfig() *modules.PostgreConfig {
	return &modules.PostgreConfig{
		Host:        "localhost",
		Port:        "5432",
		Username:    "postgres",
		Password:    "postgres",
		DBName:      "mydb",
		SSLMode:     "disable",
		ExecTimeout: 5 * time.Second,
	}
}
