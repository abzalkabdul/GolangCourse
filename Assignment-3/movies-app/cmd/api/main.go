package main

import (
	"fmt"
	"log"
	"net/http"

	"movies-app/internal/db"
	"movies-app/internal/handlers"

	"github.com/gorilla/mux"
)

func main() {
	database := db.Connect()
	defer database.Close()

	h := &handlers.Handler{DB: database}

	r := mux.NewRouter()

	r.HandleFunc("/movies", h.GetMovies).Methods("GET")
	r.HandleFunc("/movies/{id}", h.GetMovie).Methods("GET")
	r.HandleFunc("/movies", h.CreateMovie).Methods("POST")
	r.HandleFunc("/movies/{id}", h.UpdateMovie).Methods("PUT")
	r.HandleFunc("/movies/{id}", h.DeleteMovie).Methods("DELETE")

	fmt.Println("Starting the Server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
