// @title           EventBooker API
// @version         1.0
// @description     API для сервиса бронирования
// @BasePath        /

package main

import (
	"eventbooker/internal/app/booking"
	"eventbooker/internal/app/event"
	"eventbooker/internal/app/user"
	"eventbooker/internal/auth"
	"eventbooker/internal/broker/rabbit"
	"eventbooker/internal/sender"
	"eventbooker/internal/config"
	"eventbooker/internal/di"
	"eventbooker/internal/storage/postgres"
	"eventbooker/internal/web/handlers"
	wbzlog "github.com/wb-go/wbf/zlog"
	"go.uber.org/fx"
)

func main() {
	wbzlog.Init()
	app := fx.New(
		fx.Provide(
			config.NewAppConfig,
			postgres.NewPostgres,
			sender.NewEmailChannel,
			sender.NewTelegramChannel,

			func(cfg *config.AppConfig) *config.JwtConfig {
				return &cfg.JwtConfig
			},
			auth.NewJWTService,


			func(pg *postgres.Postgres) rabbit.StorageProvider {
				return pg
			},
			func(tg *sender.TelegramChannel) rabbit.TelegramProvider {
				return tg
			},
			func(mail *sender.EmailChannel) rabbit.EmailProvider {
				return mail
			},
			rabbit.NewRabbitService,

			
			func(pg *postgres.Postgres) booking.BookingStorageProvider {
				return pg
			},
			func(br *rabbit.RabbitService) booking.BookingBrokerProvider {
				return br
			},
			booking.NewBookingService,
			
			func(cfg *config.AppConfig) *config.EventConfig {
				return &cfg.EventConfig
			},
			func(pg *postgres.Postgres) event.EventStorageProvider {
				return pg
			},
			event.NewEventService,

			func(auth *auth.JWTService) user.JwtAuthProvider {
				return auth
			},
			func(pg *postgres.Postgres) user.UserStorageProvider {
				return pg
			},
			user.NewUserService,

			func(srvc *user.UserService) handlers.UserIFace {
				return srvc
			},
			handlers.NewUserHandler,

			func(srvc *event.EventService) handlers.EventIFace {
				return srvc
			},
			func(srvc *booking.BookingService) handlers.BookingIFace {
				return srvc
			},
			handlers.NewEventHandler, 

		),
		fx.Invoke(
			di.StartHTTPServer,
			di.ClosePostgresOnStop,
			di.CloseRabbitOnStop,
		),
	)

	app.Run()
}