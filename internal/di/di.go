package di

import (
	"context"
	"eventbooker/internal/broker/rabbit"
	"eventbooker/internal/config"
	"eventbooker/internal/storage/postgres"
	"eventbooker/internal/web/handlers"
	"eventbooker/internal/web/routers"
	"fmt"
	wbgin "github.com/wb-go/wbf/ginext"
	"go.uber.org/fx"
	"log"
	"net/http"
)

func StartHTTPServer(lc fx.Lifecycle, eventHandler *handlers.EventHandler, userHandler *handlers.UserHandler, config *config.AppConfig) {
	router := wbgin.New(config.GinConfig.Mode)

	router.Use(wbgin.Logger(), wbgin.Recovery())
	router.Use(func(c *wbgin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	routers.RegisterRoutes(router, userHandler, eventHandler)

	addres := fmt.Sprintf("%s:%d", config.ServerConfig.Host, config.ServerConfig.Port)
	server := &http.Server{
		Addr:    addres,
		Handler: router.Engine,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Printf("Server started")
			go func() {
				if err := server.ListenAndServe(); err != nil {
					log.Printf("ListenAndServe error: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Printf("Shutting down server...")
			return server.Close()
		},
	})
}

func ClosePostgresOnStop(lc fx.Lifecycle, postgres *postgres.Postgres) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Println("Closing Postgres connections...")
			if err := postgres.Close(); err != nil {
				log.Printf("Failed to close Postgres: %v", err)
				return err
			}
			log.Println("Postgres closed successfully")
			return nil
		},
	})
}

func CloseRabbitOnStop(lc fx.Lifecycle, rabbitCon *rabbit.RabbitService) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Println("Closing RabbitMQ client connections...")
			if err := rabbitCon.Close(); err != nil {
				log.Printf("Failed to close Postgres: %v", err)
				return err
			}
			log.Println("RabbitMQ client closed successfully")
			return nil
		},
	})
}
