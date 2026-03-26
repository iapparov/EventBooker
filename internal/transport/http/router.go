package http

import (
	_ "eventbooker/docs"
	"eventbooker/internal/transport/http/handler"
	"eventbooker/internal/transport/http/middleware"

	httpSwagger "github.com/swaggo/http-swagger"
	wbgin "github.com/wb-go/wbf/ginext"
)

// RegisterRoutes sets up all API routes.
func RegisterRoutes(engine *wbgin.Engine, userHandler *handler.UserHandler, eventHandler *handler.EventHandler, tokenValidator middleware.TokenValidator) {
	api := engine.Group("/api")

	api.GET("/swagger/*any", func(c *wbgin.Context) {
		httpSwagger.WrapHandler(c.Writer, c.Request)
	})

	// Public auth routes
	authGroup := api.Group("/auth")
	authGroup.POST("/register", userHandler.RegisterUser)
	authGroup.POST("/login", userHandler.LoginUser)
	authGroup.POST("/refresh", userHandler.RefreshToken)

	// Protected event routes
	events := api.Group("/events", middleware.Auth(tokenValidator))
	events.POST("", func(c *wbgin.Context) { eventHandler.CreateEvent(c) })
	events.GET("/:id", func(c *wbgin.Context) { eventHandler.GetEvent(c) })
	events.POST("/:id/book", func(c *wbgin.Context) { eventHandler.CreateBooking(c) })
	events.POST("/:id/confirm", func(c *wbgin.Context) { eventHandler.ConfirmBooking(c) })
}
