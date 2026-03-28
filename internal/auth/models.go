package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	IsFreeUser   bool      `json:"is_free_user"`
	CreatedAt    time.Time `json:"created_at"`
}

type RegisterRequest struct {
	Name            string `json:"name" binding:"required"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginResponse struct {
	AccessToken string      `json:"access_token"`
	TokenType   string      `json:"token_type"`
	User        UserSummary `json:"user"`
}

type UserSummary struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Repository interface {
	CreateUser(ctx context.Context, user User) (uuid.UUID, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
}

type Service interface {
	Register(ctx context.Context, name, email, password string) (uuid.UUID, error)
	Login(ctx context.Context, email, password string) (LoginResponse, error)
}
