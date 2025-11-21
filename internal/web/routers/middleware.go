package routers

import(
	wbgin "github.com/wb-go/wbf/ginext"
	"eventbooker/internal/web/handlers"
)

func AuthMiddleware(userService handlers.UserIFace) wbgin.HandlerFunc {
    return func(c *wbgin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(401, wbgin.H{"error": "missing token"})
            return
        }

        // Убираем "Bearer "
        if len(token) > 7 && token[:7] == "Bearer " {
            token = token[7:]
        }

        // Валидация через твой сервис
        payload, err := userService.ValidateTokens(token)
        if err != nil {
            c.AbortWithStatusJSON(401, wbgin.H{"error": "invalid token"})
            return
        }

        // Кладём данные из JWT в контекст
        c.Set("userId", payload.UserID)

        c.Next()
    }
}