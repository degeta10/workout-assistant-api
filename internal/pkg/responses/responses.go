package responses

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// APIResponse is your Master Template for every single response.
// Consistent keys make frontend integration seamless.
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// --- Success Helpers ---

// OK - 200 (Standard success, fetching data)
func OK(c *gin.Context, message string, data ...interface{}) {
	var payload interface{}
	if len(data) > 0 {
		payload = data[0]
	}
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: message,
		Data:    payload,
	})
}

// Created - 201 (Resource created successfully)
func Created(c *gin.Context, message string, data ...interface{}) {
	var payload interface{}
	if len(data) > 0 {
		payload = data[0]
	}
	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: message,
		Data:    payload,
	})
}

// --- Error Helpers ---

// BadRequest - 400 (Client-side syntax/logic error)
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, APIResponse{
		Success: false,
		Message: message,
	})
}

// Unauthorized - 401 (Unauthenticated or invalid credentials)
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, APIResponse{
		Success: false,
		Message: message,
	})
}

// Forbidden - 403 (Authenticated but not authorized for this action)
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, APIResponse{
		Success: false,
		Message: message,
	})
}

// NotFound - 404 (Resource doesn't exist)
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, APIResponse{
		Success: false,
		Message: message,
	})
}

// InternalError - 500 (Server-side crash/DB failure)
func InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, APIResponse{
		Success: false,
		Message: message,
	})
}

// --- The Special "Validation" Helper ---

// ValidationError - 422 (The Laravel standard for failed validation)
func ValidationError(c *gin.Context, err error) {
	fieldErrors := make(map[string]string)
	var ve validator.ValidationErrors

	if errors.As(err, &ve) {
		for _, f := range ve {
			// Converts "ConfirmPassword" -> "confirm_password"
			field := toSnakeCase(f.Field())
			fieldErrors[field] = translateTag(f)
		}
	} else {
		fieldErrors["error"] = "Invalid request format"
	}

	c.JSON(http.StatusUnprocessableEntity, APIResponse{
		Success: false,
		Message: "The given data was invalid.",
		Errors:  fieldErrors,
	})
}

// --- Utilities ---

func translateTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required."
	case "email":
		return "Must be a valid email address."
	case "min":
		return "Must be at least " + fe.Param() + " characters."
	case "eqfield":
		return "Does not match the " + strings.ToLower(fe.Param()) + " field."
	}
	return "Invalid value."
}

func toSnakeCase(str string) string {
	var result strings.Builder
	for i, r := range str {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
