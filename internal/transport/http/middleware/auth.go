package middleware

import (
	"eventbooker/internal/auth"

	wbgin "github.com/wb-go/wbf/ginext"
)

// TokenValidator defines the interface for validating JWT tokens.
type TokenValidator interface {
	ValidateToken(tokenStr string) (*auth.Payload, error)
}

// Auth returns a middleware that validates JWT tokens.
func Auth(validator TokenValidator) wbgin.HandlerFunc {
	return func(c *wbgin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(401, wbgin.H{"error": "missing token"})
			return
		}

		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		payload, err := validator.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(401, wbgin.H{"error": "invalid token"})
			return
		}

		c.Set("userId", payload.UserID)
		c.Next()
	}
}
