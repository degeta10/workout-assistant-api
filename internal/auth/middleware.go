package auth

import (
	"context"
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

// RequireAuth is the middleware that protects your private routes
func RequireAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Extract the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			responses.Unauthorized(c, "Authorization header is missing")
			c.Abort()
			return
		}

		// 2. Check the Bearer prefix
		parts := strings.Fields(authHeader)
		if len(parts) != 2 || parts[0] != "Bearer" {
			responses.Unauthorized(c, "Invalid authorization format. Use 'Bearer <token>'")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 3. Parse and Validate the JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Ensure the signing method is what we expect
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			responses.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// 4. Extract the User ID (sub) and attach it to the Gin Context
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			userIDStr, ok := claims["sub"].(string)
			if !ok || userIDStr == "" {
				responses.Unauthorized(c, "Invalid user ID in token")
				c.Abort()
				return
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				responses.Unauthorized(c, "Invalid user ID in token")
				c.Abort()
				return
			}

			// Attach UUID to the request context for downstream service layers.
			ctx := context.WithValue(c.Request.Context(), userIDContextKey, userID)
			c.Request = c.Request.WithContext(ctx)
			c.Next() // Pass control to the actual handler (like /me)
		} else {
			responses.Unauthorized(c, "Invalid token claims")
			c.Abort()
		}
	}
}
