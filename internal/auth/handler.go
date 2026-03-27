package auth

import (
	"net/http"
	"os"
	"time"

	"github.com/degeta10/workout-assistant-api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Register godoc
// @Summary Register a new user
// @Description Create a user with name, email, and password
// @Tags auth
// @Accept json
// @Produce json
// @Param user body models.RegisterRequest true "User Registration Data"
// @Success 201 {object} map[string]interface{}
// @Router /register [post]
func Register(c *gin.Context) {
	var input models.RegisterRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

	id, err := CreateUser(input.Name, input.Email, string(hash))
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already in use"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user_id": id, "message": "Success"})
}

// Login godoc
// @Summary Authenticate user and return JWT
// @Description Login with email and password to receive a JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body models.AuthRequest true "User Credentials"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /login [post]
func Login(c *gin.Context) {
	var input models.AuthRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := GetUserByEmail(input.Email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID.String(),
		"exp": time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}
