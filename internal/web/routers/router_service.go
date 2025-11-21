package routers

import (
	httpSwagger "github.com/swaggo/http-swagger"
	wbgin "github.com/wb-go/wbf/ginext"
	"eventbooker/internal/web/handlers"
	// _ "imageProcessor/docs"
)

func RegisterRoutes(engine *wbgin.Engine, userhandler *handlers.UserHandler, eventhandler *handlers.EventHandler) {
	api := engine.Group("/api")
	api.GET("/swagger/*any", func(c *wbgin.Context) {
		httpSwagger.WrapHandler(c.Writer, c.Request)
	})

	// публичные маршруты авторизации
	auth := api.Group("/auth")
	auth.POST("/register", userhandler.RegisterUser)
	auth.POST("/login", userhandler.LoginUser)
	auth.POST("/refresh", userhandler.RefreshToken)

	// защищённая группа ивентов
	events := api.Group("/events", AuthMiddleware(userhandler.Service))
	{
		events.POST("", func(c *wbgin.Context) { eventhandler.CreateEvent(c) })
		events.GET("/:id", func(c *wbgin.Context) { eventhandler.GetEvent(c) })

		events.POST("/:id/book", func(c *wbgin.Context) { eventhandler.CreateBooking(c) })
		events.POST("/:id/confirm", func(c *wbgin.Context) { eventhandler.ConfirmBooking(c) })
	}
}