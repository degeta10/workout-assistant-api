package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/degeta10/workout-assistant-api/internal/pkg/responses"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// contextKey is an unexported type for context keys in the auth package,
// preventing collisions with keys from other packages.
type contextKey string

const userIDContextKey contextKey = "userID"

var (
	errMissingAuthorization = errors.New("authorization header is missing")
	errInvalidFormat        = errors.New("invalid authorization format")
	errInvalidToken         = errors.New("invalid or expired token")
	errInvalidClaims        = errors.New("invalid token claims")
	errInvalidUserID        = errors.New("invalid user id in token")
)

// RequireAuth is the middleware that protects your private routes
func RequireAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := extractBearerToken(c.GetHeader("Authorization"))
		if err != nil {
			writeUnauthorized(c, err)
			return
		}

		token, err := parseAndValidateToken(tokenString, jwtSecret)
		if err != nil {
			writeUnauthorized(c, err)
			return
		}

		userID, err := extractUserID(token)
		if err != nil {
			writeUnauthorized(c, err)
			return
		}

		ctx := context.WithValue(c.Request.Context(), userIDContextKey, userID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func extractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errMissingAuthorization
	}

	parts := strings.Fields(authHeader)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errInvalidFormat
	}

	return parts[1], nil
}

func parseAndValidateToken(tokenString, jwtSecret string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errInvalidToken
	}

	return token, nil
}

func extractUserID(token *jwt.Token) (uuid.UUID, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, errInvalidClaims
	}

	userIDStr, ok := claims["sub"].(string)
	if !ok || userIDStr == "" {
		return uuid.Nil, errInvalidUserID
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, errInvalidUserID
	}

	return userID, nil
}

func writeUnauthorized(c *gin.Context, err error) {
	message := "Invalid or expired token"
	switch {
	case errors.Is(err, errMissingAuthorization):
		message = "Authorization header is missing"
	case errors.Is(err, errInvalidFormat):
		message = "Invalid authorization format. Use 'Bearer <token>'"
	case errors.Is(err, errInvalidClaims):
		message = "Invalid token claims"
	case errors.Is(err, errInvalidUserID):
		message = "Invalid user ID in token"
	}

	responses.Unauthorized(c, message)
	c.Abort()
}
