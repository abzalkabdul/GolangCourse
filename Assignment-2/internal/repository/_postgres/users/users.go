package users

import (
	"assignment-2/internal/repository/_postgres"
	"assignment-2/pkg/modules"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

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

// GetUsers returns all users from the database.
func (r *Repository) GetUsers() ([]modules.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.executionTimeout)
	defer cancel()

	var users []modules.User
	err := r.db.DB.SelectContext(ctx, &users,
		"SELECT id, name, email, age, created_at FROM users ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("GetUsers: %w", err)
	}
	return users, nil
}

// GetUserByID fetches a single user by ID.
// Returns nil + descriptive error if not found.
func (r *Repository) GetUserByID(id int) (*modules.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.executionTimeout)
	defer cancel()

	var user modules.User
	err := r.db.DB.GetContext(ctx, &user,
		"SELECT id, name, email, age, created_at FROM users WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user with id=%d not found", id)
		}
		return nil, fmt.Errorf("GetUserByID(id=%d): %w", id, err)
	}
	return &user, nil
}

// CreateUser inserts a new user and returns the newly generated ID.
func (r *Repository) CreateUser(req modules.CreateUserRequest) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.executionTimeout)
	defer cancel()

	var id int
	err := r.db.DB.QueryRowContext(ctx,
		`INSERT INTO users (name, email, age) VALUES ($1, $2, $3) RETURNING id`,
		req.Name, req.Email, req.Age,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("CreateUser: %w", err)
	}
	return id, nil
}

// UpdateUser updates name/email/age for the given ID.
// Returns a custom error if no row was found.
func (r *Repository) UpdateUser(id int, req modules.UpdateUserRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.executionTimeout)
	defer cancel()

	result, err := r.db.DB.ExecContext(ctx,
		`UPDATE users SET name=$1, email=$2, age=$3 WHERE id=$4`,
		req.Name, req.Email, req.Age, id,
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

// DeleteUser removes a user by ID and returns the number of rows affected.
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
