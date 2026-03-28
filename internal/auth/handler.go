package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	group.POST("/register", h.Register)
	group.POST("/login", h.Login)
}

// Register godoc
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "User Registration Data"
// @Success 201 {object} map[string]interface{}
// @Router /register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.svc.Register(c.Request.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already in use"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user_id": id.String(), "message": "Success"})
}

// Login godoc
// @Summary Authenticate user and return JWT
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "User Credentials"
// @Success 200 {object} map[string]interface{}
// @Router /login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
