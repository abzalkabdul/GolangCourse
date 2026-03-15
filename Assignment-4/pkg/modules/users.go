package modules

import "time"

// User is the core domain model - extended with gender and birth_date for Assignment 3.
type User struct {
	ID        int       `db:"id"         json:"id"`
	Name      string    `db:"name"       json:"name"`
	Email     string    `db:"email"      json:"email"`
	Age       int       `db:"age"        json:"age"`
	Gender    string    `db:"gender"     json:"gender"`
	BirthDate time.Time `db:"birth_date" json:"birth_date"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// PaginatedResponse wraps a page of users with pagination metadata.
type PaginatedResponse struct {
	Data       []User `json:"data"`
	TotalCount int    `json:"totalCount"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
}

// UserFilter holds all optional query-string filters + pagination + sorting.
type UserFilter struct {
	// Pagination
	Page     int
	PageSize int
	// Sorting
	OrderBy  string // column name, whitelisted before use
	OrderDir string // "asc" or "desc"
	ID        *int
	Name      string
	Email     string
	Gender    string
	BirthDate string // YYYY-MM-DD, compared with DATE trunc
}

type CreateUserRequest struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	Gender    string    `json:"gender"`
	BirthDate time.Time `json:"birth_date"`
}

type UpdateUserRequest struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	Gender    string    `json:"gender"`
	BirthDate time.Time `json:"birth_date"`
}
