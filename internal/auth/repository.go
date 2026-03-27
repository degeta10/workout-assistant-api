package auth

import (
	"github.com/degeta10/workout-assistant-api/internal/database"
	"github.com/degeta10/workout-assistant-api/internal/models"
	"github.com/google/uuid"
)

func CreateUser(name, email, hash string) (uuid.UUID, error) {
	query := `INSERT INTO users (name, email, password_hash) 
              VALUES ($1, $2, $3) RETURNING id`

	var id uuid.UUID
	err := database.DB.QueryRow(query, name, email, hash).Scan(&id)
	return id, err
}

func GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, name, email, password_hash, is_free_user, created_at FROM users WHERE email = $1`
	err := database.DB.QueryRow(query, email).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.IsFreeUser, &user.CreatedAt)
	return user, err
}
