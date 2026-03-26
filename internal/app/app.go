package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"eventbooker/internal/auth"
	"eventbooker/internal/broker/rabbit"
	"eventbooker/internal/config"
	"eventbooker/internal/notification"
	"eventbooker/internal/repository/postgres"
	"eventbooker/internal/service"
	httpTransport "eventbooker/internal/transport/http"
	"eventbooker/internal/transport/http/handler"

	wbgin "github.com/wb-go/wbf/ginext"
	wbzlog "github.com/wb-go/wbf/zlog"
)

// App is the main application container.
type App struct {
	cfg      *config.AppConfig
	server   *http.Server
	postgres *postgres.Repository
	broker   *rabbit.Broker
}

// New initializes all dependencies and creates the App.
func New(cfg *config.AppConfig) (*App, error) {
	// Infrastructure
	pg, err := postgres.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("postgres: %w", err)
	}

	emailSender := notification.NewEmailSender(cfg)
	telegramSender := notification.NewTelegramSender(cfg)

	broker, err := rabbit.NewBroker(cfg, pg, emailSender, telegramSender)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq: %w", err)
	}

	// Auth
	jwtService := auth.NewService(&cfg.JWT)

	// Services
	bookingSvc := service.NewBookingService(pg, broker)
	eventSvc := service.NewEventService(pg, &cfg.Event)
	userSvc := service.NewUserService(pg, jwtService, cfg)

	// Handlers
	eventHandler := handler.NewEventHandler(eventSvc, bookingSvc)
	userHandler := handler.NewUserHandler(userSvc)

	// Router
	router := wbgin.New(cfg.Gin.Mode)
	router.Use(wbgin.Logger(), wbgin.Recovery())
	router.Use(corsMiddleware())

	httpTransport.RegisterRoutes(router, userHandler, eventHandler, userSvc)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: router.Engine,
	}

	return &App{
		cfg:      cfg,
		server:   server,
		postgres: pg,
		broker:   broker,
	}, nil
}

// Run starts the HTTP server and blocks until a shutdown signal is received.
func (a *App) Run() {
	// Start server
	go func() {
		wbzlog.Logger.Info().Msgf("server started on %s", a.server.Addr)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			wbzlog.Logger.Fatal().Err(err).Msg("server failed")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	wbzlog.Logger.Info().Msg("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		wbzlog.Logger.Error().Err(err).Msg("server forced to shutdown")
	}

	a.stop()

	wbzlog.Logger.Info().Msg("server exited properly")
}

func (a *App) stop() {
	if err := a.postgres.Close(); err != nil {
		log.Printf("failed to close Postgres: %v", err)
	} else {
		log.Println("Postgres closed successfully")
	}

	if err := a.broker.Close(); err != nil {
		log.Printf("failed to close RabbitMQ: %v", err)
	} else {
		log.Println("RabbitMQ closed successfully")
	}
}

func corsMiddleware() wbgin.HandlerFunc {
	return func(c *wbgin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
