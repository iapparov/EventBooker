package rabbit

import (
	"context"
	"eventbooker/internal/config"
	"eventbooker/internal/domain/booking"
	"fmt"
	"time"
	"github.com/rabbitmq/amqp091-go"
	wbrabbit "github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/retry"
	wbzlog "github.com/wb-go/wbf/zlog"
)

type RabbitService struct {
	client *wbrabbit.RabbitClient
	publisher *wbrabbit.Publisher
	consumer *wbrabbit.Consumer
}

type StorageProvider interface {
	CancelBooking(ctx context.Context, bookingid, event_id string) error
	GetBookingStatus(ctx context.Context, id string) (booking.BookingStatus, error)
}

type TelegramProvider interface {
	Send(tg string, EventName string, Persons int) error
}

type EmailProvider interface {
	Send(email string, EventName string, Persons int) error
}

func NewRabbitService(cfg *config.AppConfig, repo StorageProvider, email EmailProvider, tg TelegramProvider) (*RabbitService, error) {
	rabbitDSN := fmt.Sprintf(
		"amqp://%s:%s@%s:%d/",
		cfg.RabbitmqConfig.User,
		cfg.RabbitmqConfig.Password,
		cfg.RabbitmqConfig.Host,
		cfg.RabbitmqConfig.Port,
	)
	rabbitConfig := wbrabbit.ClientConfig{
		URL: rabbitDSN,
		ConnectionName: cfg.RabbitmqConfig.ConnectionName,
		ConnectTimeout: time.Duration(cfg.RabbitmqConfig.ConnectionTimeout) * time.Second,
    	Heartbeat:      time.Duration(cfg.RabbitmqConfig.Heartbeat) * time.Second, // стабильное поддержание коннекта
		PublishRetry: retry.Strategy{
			Attempts: cfg.RetrysConfig.Attempts,
			Delay: cfg.RetrysConfig.Delay,
			Backoff: cfg.RetrysConfig.Backoffs,
		},
		ConsumeRetry: retry.Strategy{
			Attempts: cfg.RetrysConfig.Attempts,
			Delay: cfg.RetrysConfig.Delay,
			Backoff: cfg.RetrysConfig.Backoffs,
		},
	}

	client, err := wbrabbit.NewClient(rabbitConfig)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("bad rabbitmq client connection")
		return nil, err
	}
	err = declareInfrastructure(client, cfg)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("bad rabbitmq client declareInfrastructure")
		return nil, err
	}
	publisher := wbrabbit.NewPublisher(client, "booking.delay.exchange", "application/json")
	consumer := wbrabbit.NewConsumer(client, wbrabbit.ConsumerConfig{
		Queue:         "expired.queue",
		ConsumerTag:   "booking-expired-worker",
		AutoAck:       false,
		PrefetchCount: 10,
		Workers:       5,
	}, BookingExpiredHandler(repo, email, tg))

	go consumer.Start(context.Background())
	return &RabbitService{
		client: client,
		publisher: publisher,
		consumer: consumer,
	}, nil
}

func (s *RabbitService) Close() error {
	return s.client.Close()
}


func declareInfrastructure(client *wbrabbit.RabbitClient, cfg *config.AppConfig) error {
	// 1. DLX — когда TTL истёк
	if err := client.DeclareExchange(
		"booking.dlx.exchange",
		"direct",
		true,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	//
	// 2. Очередь, куда падают ТОЛЬКО просроченные сообщения
	//
	if err := client.DeclareQueue(
		"expired.queue",          // queue name
		"booking.dlx.exchange",   // exchange
		"booking.expired",        // routing key
		true,                     // durable
		false,                    // autoDelete
		true,                     // exchange durable
		nil,                      // args
	); err != nil {
		return err
	}

    //
    // 3. Delay exchange — publisher шлёт сюда
    //
    if err := client.DeclareExchange(
        "booking.delay.exchange",
        "direct",
        true,
        false,
        false,
        nil,
    ); err != nil {
        return err
    }

//
    // 4. Delay queues — динамически на основе SupportedTTLs
    //
    for ttlMinutes := range cfg.EventConfig.SupportedTTLs {

        ttlMs := ttlMinutes * 60 * 1000 // Rabbit принимает TTL в миллисекундах

        queueName := fmt.Sprintf("delay_%d.queue", ttlMinutes)
        routingKey := fmt.Sprintf("delay_%d", ttlMinutes)

        args := amqp091.Table{
            "x-message-ttl":             ttlMs,
            "x-dead-letter-exchange":    "booking.dlx.exchange",
            "x-dead-letter-routing-key": "booking.expired",
        }

        // объявляем delay очередь (с TTL + DLX)
        if err := client.DeclareQueue(
            queueName,
            "booking.delay.exchange",
            routingKey,
            true,  // durable
            false, // autoDelete
            true,  // exchange durable
            args,
        ); err != nil {
            return err
        }
    }

	return nil
}