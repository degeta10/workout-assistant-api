package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateUser(ctx context.Context, user User) (uuid.UUID, error) {
	if r.db == nil {
		return uuid.Nil, fmt.Errorf("create user: database not initialized")
	}

	const query = `INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id`

	var id uuid.UUID
	err := r.db.QueryRowContext(ctx, query, user.Name, user.Email, user.PasswordHash).Scan(&id)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return uuid.Nil, ErrEmailAlreadyExists
		}
		return uuid.Nil, fmt.Errorf("insert user: %w", err)
	}

	return id, nil
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	if r.db == nil {
		return nil, fmt.Errorf("get user by email: database not initialized")
	}

	const query = `SELECT id, name, email, password_hash, is_free_user, created_at
		FROM users
		WHERE email = $1`

	user := &User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.IsFreeUser,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("select user by email: %w", err)
	}

	return user, nil
}
