package routers

import (
	"eventbooker/internal/web/handlers"
	wbgin "github.com/wb-go/wbf/ginext"
)

func AuthMiddleware(userService handlers.UserIFace) wbgin.HandlerFunc {
	return func(c *wbgin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(401, wbgin.H{"error": "missing token"})
			return
		}

		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		payload, err := userService.ValidateTokens(token)
		if err != nil {
			c.AbortWithStatusJSON(401, wbgin.H{"error": "invalid token"})
			return
		}

		c.Set("userId", payload.UserID)

		c.Next()
	}
}
