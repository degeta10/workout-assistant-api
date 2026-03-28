package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo      Repository
	jwtSecret string
}

func NewService(repo Repository, jwtSecret string) Service {
	return &AuthService{repo: repo, jwtSecret: jwtSecret}
}

func (s *AuthService) Register(ctx context.Context, name, email, password string) (uuid.UUID, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, fmt.Errorf("hash password: %w", err)
	}

	id, err := s.repo.CreateUser(ctx, User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("create user: %w", err)
	}

	return id, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (LoginResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if err == ErrInvalidCredentials {
			return LoginResponse{}, err
		}
		return LoginResponse{}, fmt.Errorf("get user by email: %w", err)
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return LoginResponse{}, ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID.String(),
		"exp": time.Now().Add(72 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return LoginResponse{}, fmt.Errorf("sign jwt token: %w", err)
	}

	return LoginResponse{
		AccessToken: tokenString,
		TokenType:   "Bearer",
		User: UserSummary{
			Name:  user.Name,
			Email: user.Email,
		},
	}, nil
}

func (s *AuthService) Me(ctx context.Context) (UserSummary, error) {
	userIDVal := ctx.Value("userID")
	if userIDVal == nil {
		return UserSummary{}, ErrUserIDMissing
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		return UserSummary{}, fmt.Errorf("invalid user id type in context")
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return UserSummary{}, fmt.Errorf("get user by id: %w", err)
	}

	return UserSummary{
		Name:  user.Name,
		Email: user.Email,
	}, nil
}
