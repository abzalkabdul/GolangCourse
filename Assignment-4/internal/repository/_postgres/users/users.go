package users

import (
	"assignment-4/internal/repository/_postgres"
	"assignment-4/pkg/modules"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// allowedColumns is the whitelist for ORDER BY / filter columns (SQL-injection protection).
var allowedColumns = map[string]string{
	"id":         "id",
	"name":       "name",
	"email":      "email",
	"gender":     "gender",
	"birth_date": "birth_date",
	"age":        "age",
	"created_at": "created_at",
}

type Repository struct {
	db               *_postgres.Dialect
	executionTimeout time.Duration
}

func NewUserRepository(db *_postgres.Dialect) *Repository {
	return &Repository{
		db:               db,
		executionTimeout: time.Second * 5,
	}
}

func (r *Repository) GetUsers() ([]modules.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.executionTimeout)
	defer cancel()

	var users []modules.User
	err := r.db.DB.SelectContext(ctx, &users,
		"SELECT id, name, email, age, gender, birth_date, created_at FROM users ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("GetUsers: %w", err)
	}
	return users, nil
}

func (r *Repository) GetUserByID(id int) (*modules.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.executionTimeout)
	defer cancel()

	var user modules.User
	err := r.db.DB.GetContext(ctx, &user,
		"SELECT id, name, email, age, gender, birth_date, created_at FROM users WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user with id=%d not found", id)
		}
		return nil, fmt.Errorf("GetUserByID(id=%d): %w", id, err)
	}
	return &user, nil
}

func (r *Repository) CreateUser(req modules.CreateUserRequest) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.executionTimeout)
	defer cancel()

	var id int
	err := r.db.DB.QueryRowContext(ctx,
		`INSERT INTO users (name, email, age, gender, birth_date)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		req.Name, req.Email, req.Age, req.Gender, req.BirthDate,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("CreateUser: %w", err)
	}
	return id, nil
}

func (r *Repository) UpdateUser(id int, req modules.UpdateUserRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.executionTimeout)
	defer cancel()

	result, err := r.db.DB.ExecContext(ctx,
		`UPDATE users SET name=$1, email=$2, age=$3, gender=$4, birth_date=$5 WHERE id=$6`,
		req.Name, req.Email, req.Age, req.Gender, req.BirthDate, id,
	)
	if err != nil {
		return fmt.Errorf("UpdateUser(id=%d): %w", id, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("UpdateUser RowsAffected(id=%d): %w", id, err)
	}
	if affected == 0 {
		return fmt.Errorf("user with id=%d does not exist, nothing was updated", id)
	}
	return nil
}

func (r *Repository) DeleteUser(id int) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.executionTimeout)
	defer cancel()

	result, err := r.db.DB.ExecContext(ctx, "DELETE FROM users WHERE id=$1", id)
	if err != nil {
		return 0, fmt.Errorf("DeleteUser(id=%d): %w", id, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("DeleteUser RowsAffected(id=%d): %w", id, err)
	}
	if affected == 0 {
		return 0, fmt.Errorf("user with id=%d does not exist, nothing was deleted", id)
	}
	return affected, nil
}

func (r *Repository) GetPaginatedUsers(f modules.UserFilter) (modules.PaginatedResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.executionTimeout)
	defer cancel()

	args := []any{}
	conditions := []string{}
	argIdx := 1

	if f.ID != nil {
		conditions = append(conditions, fmt.Sprintf("id = $%d", argIdx))
		args = append(args, *f.ID)
		argIdx++
	}
	if f.Name != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIdx))
		args = append(args, "%"+f.Name+"%")
		argIdx++
	}
	if f.Email != "" {
		conditions = append(conditions, fmt.Sprintf("email ILIKE $%d", argIdx))
		args = append(args, "%"+f.Email+"%")
		argIdx++
	}
	if f.Gender != "" {
		conditions = append(conditions, fmt.Sprintf("gender = $%d", argIdx))
		args = append(args, f.Gender)
		argIdx++
	}
	if f.BirthDate != "" {
		conditions = append(conditions, fmt.Sprintf("DATE(birth_date) = $%d", argIdx))
		args = append(args, f.BirthDate)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	// -- COUNT total matching rows --
	var totalCount int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users %s", where)
	err := r.db.DB.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return modules.PaginatedResponse{}, fmt.Errorf("GetPaginatedUsers count: %w", err)
	}

	// -- ORDER BY (whitelisted) --
	orderCol := "id" // default
	if col, ok := allowedColumns[f.OrderBy]; ok {
		orderCol = col
	}
	orderDir := "ASC"
	if strings.ToLower(f.OrderDir) == "desc" {
		orderDir = "DESC"
	}

	// -- LIMIT / OFFSET --
	pageSize := f.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	page := f.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	// Add pagination args
	query := fmt.Sprintf(
		`SELECT id, name, email, age, gender, birth_date, created_at
		 FROM users %s
		 ORDER BY %s %s
		 LIMIT $%d OFFSET $%d`,
		where, orderCol, orderDir, argIdx, argIdx+1,
	)
	args = append(args, pageSize, offset)

	rows, err := r.db.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return modules.PaginatedResponse{}, fmt.Errorf("GetPaginatedUsers query: %w", err)
	}
	defer rows.Close()

	var users []modules.User
	for rows.Next() {
		var u modules.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Age, &u.Gender, &u.BirthDate, &u.CreatedAt); err != nil {
			return modules.PaginatedResponse{}, fmt.Errorf("GetPaginatedUsers scan: %w", err)
		}
		users = append(users, u)
	}
	if users == nil {
		users = []modules.User{}
	}

	return modules.PaginatedResponse{
		Data:       users,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// GetCommonFriends returns users who are friends with BOTH user1 and user2.
// Uses a single JOIN query to avoid the N+1 problem.
func (r *Repository) GetCommonFriends(userID1, userID2 int) ([]modules.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.executionTimeout)
	defer cancel()

	query := `
		SELECT u.id, u.name, u.email, u.age, u.gender, u.birth_date, u.created_at
		FROM users u
		JOIN user_friends f1 ON f1.friend_id = u.id AND f1.user_id = $1
		JOIN user_friends f2 ON f2.friend_id = u.id AND f2.user_id = $2
		ORDER BY u.id
	`

	rows, err := r.db.DB.QueryContext(ctx, query, userID1, userID2)
	if err != nil {
		return nil, fmt.Errorf("GetCommonFriends: %w", err)
	}
	defer rows.Close()

	var users []modules.User
	for rows.Next() {
		var u modules.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Age, &u.Gender, &u.BirthDate, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("GetCommonFriends scan: %w", err)
		}
		users = append(users, u)
	}
	if users == nil {
		users = []modules.User{}
	}
	return users, nil
}
