package auth

import (
	"errors"

	"github.com/degeta10/workout-assistant-api/internal/pkg/apperrors"
	"github.com/degeta10/workout-assistant-api/internal/pkg/responses"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterPublicRoutes(group *gin.RouterGroup) {
	group.POST("/register", h.Register)
	group.POST("/login", h.Login)
}

func (h *Handler) RegisterProtectedRoutes(group *gin.RouterGroup) {
	group.GET("/me", h.Me)
}

// Register godoc
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "User Registration Data"
// @Success 201 {object} responses.APIResponse
// @Failure 409 {object} responses.APIResponse
// @Failure 422 {object} responses.APIResponse
// @Failure 500 {object} responses.APIResponse
// @Router /register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.Validation(err))
		return
	}

	_, err := h.svc.Register(c.Request.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			_ = c.Error(apperrors.Conflict("Email already in use", err))
			return
		}
		_ = c.Error(apperrors.Internal("Failed to create user", err))
		return
	}

	responses.Created(c, "User registered successfully")
}

// Login godoc
// @Summary Authenticate user and return JWT
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "User Credentials"
// @Success 200 {object} responses.APIResponse
// @Failure 401 {object} responses.APIResponse
// @Failure 422 {object} responses.APIResponse
// @Failure 500 {object} responses.APIResponse
// @Router /login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.Validation(err))
		return
	}

	data, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			_ = c.Error(apperrors.Unauthorized("Invalid credentials", err))
			return
		}
		_ = c.Error(apperrors.Internal("Failed to authenticate", err))
		return
	}

	responses.OK(c, "Login successful", data)
}

// Me godoc
// @Summary Get current authenticated user info
// @Tags auth
// @Produce json
// @Success 200 {object} responses.APIResponse
// @Failure 401 {object} responses.APIResponse
// @Failure 500 {object} responses.APIResponse
// @Router /me [get]
func (h *Handler) Me(c *gin.Context) {
	user, err := h.svc.Me(c.Request.Context())
	if err != nil {
		if errors.Is(err, ErrUserIDMissing) || errors.Is(err, ErrUserNotFound) {
			_ = c.Error(apperrors.Unauthorized("Invalid credentials", err))
			return
		}
		_ = c.Error(apperrors.Internal("Failed to fetch user", err))
		return
	}

	responses.OK(c, "User fetched successfully", user)
}
