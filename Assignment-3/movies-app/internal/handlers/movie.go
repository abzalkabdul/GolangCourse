package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"movies-app/internal/models"

	"github.com/gorilla/mux"
)

type Handler struct {
	DB *sql.DB
}

// GET /movies
func (h *Handler) GetMovies(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`
		SELECT m.id, m.title, m.genre, m.budget,
		       t.id, t.name, t.role
		FROM movies m
		LEFT JOIN technicians t ON t.movie_id = m.id
		ORDER BY m.id
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	moviesMap := map[int]*models.Movie{}
	var order []int

	for rows.Next() {
		var m models.Movie
		var tID sql.NullInt64
		var tName, tRole sql.NullString

		if err := rows.Scan(&m.ID, &m.Title, &m.Genre, &m.Budget, &tID, &tName, &tRole); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, exists := moviesMap[m.ID]; !exists {
			moviesMap[m.ID] = &m
			order = append(order, m.ID)
		}

		if tID.Valid {
			moviesMap[m.ID].Technicians = append(moviesMap[m.ID].Technicians, models.Technician{
				ID:   int(tID.Int64),
				Name: tName.String,
				Role: tRole.String,
			})
		}
	}

	result := make([]*models.Movie, 0, len(order))
	for _, id := range order {
		result = append(result, moviesMap[id])
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GET /movies/{id}
func (h *Handler) GetMovie(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	rows, err := h.DB.Query(`
		SELECT m.id, m.title, m.genre, m.budget,
		       t.id, t.name, t.role
		FROM movies m
		LEFT JOIN technicians t ON t.movie_id = m.id
		WHERE m.id = $1
	`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var movie *models.Movie

	for rows.Next() {
		var m models.Movie
		var tID sql.NullInt64
		var tName, tRole sql.NullString

		if err := rows.Scan(&m.ID, &m.Title, &m.Genre, &m.Budget, &tID, &tName, &tRole); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if movie == nil {
			movie = &m
		}
		if tID.Valid {
			movie.Technicians = append(movie.Technicians, models.Technician{
				ID:   int(tID.Int64),
				Name: tName.String,
				Role: tRole.String,
			})
		}
	}

	if movie == nil {
		http.Error(w, "movie not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movie)
}

// POST /movies
func (h *Handler) CreateMovie(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title       string              `json:"title"`
		Genre       string              `json:"genre"`
		Budget      int64               `json:"budget"`
		Technicians []models.Technician `json:"technicians"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var movieID int
	err = tx.QueryRow(`INSERT INTO movies (title, genre, budget) VALUES ($1, $2, $3) RETURNING id`,
		input.Title, input.Genre, input.Budget).Scan(&movieID)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, t := range input.Technicians {
		_, err = tx.Exec(`INSERT INTO technicians (movie_id, name, role) VALUES ($1, $2, $3)`,
			movieID, t.Name, t.Role)
		if err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	tx.Commit()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"id": movieID, "message": "movie created"})
}

// PUT /movies/{id}
func (h *Handler) UpdateMovie(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	var input struct {
		Title  string `json:"title"`
		Genre  string `json:"genre"`
		Budget int64  `json:"budget"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec(`UPDATE movies SET title=$1, genre=$2, budget=$3 WHERE id=$4`,
		input.Title, input.Genre, input.Budget, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.Error(w, "movie not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "movie updated"})
}

// DELETE /movies/{id}
func (h *Handler) DeleteMovie(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	result, err := h.DB.Exec(`DELETE FROM movies WHERE id=$1`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.Error(w, "movie not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "movie deleted"})
}
