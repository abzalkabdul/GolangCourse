package handler

import (
	"assignment-4/internal/usecase"
	"assignment-4/pkg/modules"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type UserHandler struct {
	uc *usecase.UserUsecase
}

func NewUserHandler(uc *usecase.UserUsecase) *UserHandler {
	return &UserHandler{uc: uc}
}

// writeJSON sends a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError sends a JSON error envelope.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// GetUsers handles GET /users
func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.uc.GetUsers()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	user, err := h.uc.GetUserByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req modules.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id, err := h.uc.CreateUser(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int{"id": id})
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var req modules.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.uc.UpdateUser(id, req); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "user updated successfully"})
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	affected, err := h.uc.DeleteUser(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"message":       "user deleted successfully",
		"rows_affected": affected,
	})
}

func (h *UserHandler) GetPaginatedUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	f := modules.UserFilter{
		Page:      queryInt(q.Get("page"), 1),
		PageSize:  queryInt(q.Get("page_size"), 10),
		OrderBy:   q.Get("order_by"),
		OrderDir:  q.Get("order_dir"),
		Name:      q.Get("name"),
		Email:     q.Get("email"),
		Gender:    q.Get("gender"),
		BirthDate: q.Get("birth_date"),
	}

	if idStr := q.Get("id"); idStr != "" {
		if idVal, err := strconv.Atoi(idStr); err == nil {
			f.ID = &idVal
		} else {
			writeError(w, http.StatusBadRequest, "id must be an integer")
			return
		}
	}

	result, err := h.uc.GetPaginatedUsers(f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// GetCommonFriends handles GET /users/common-friends?user1=1&user2=2
func (h *UserHandler) GetCommonFriends(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	user1, err1 := strconv.Atoi(q.Get("user1"))
	user2, err2 := strconv.Atoi(q.Get("user2"))
	if err1 != nil || err2 != nil {
		writeError(w, http.StatusBadRequest, "user1 and user2 must be valid integer IDs")
		return
	}

	friends, err := h.uc.GetCommonFriends(user1, user2)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user1":          user1,
		"user2":          user2,
		"common_friends": friends,
		"count":          len(friends),
	})
}

// extractID pulls the last path segment and parses it as int. /users/42 -> 42
func extractID(path string) (int, error) {
	parts := strings.Split(strings.TrimRight(path, "/"), "/")
	return strconv.Atoi(parts[len(parts)-1])
}

// queryInt parses a query string value as int, returning fallback on failure.
func queryInt(s string, fallback int) int {
	if v, err := strconv.Atoi(s); err == nil && v > 0 {
		return v
	}
	return fallback
}
