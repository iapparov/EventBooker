package rabbit

import (
	"context"
	"fmt"
	"os"
	"time"

	"eventbooker/internal/config"
	"eventbooker/internal/domain/booking"

	"github.com/rabbitmq/amqp091-go"
	wbrabbit "github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/retry"
	wbzlog "github.com/wb-go/wbf/zlog"
)

// Broker wraps a RabbitMQ client with publisher and consumer.
type Broker struct {
	client    *wbrabbit.RabbitClient
	publisher *wbrabbit.Publisher
	consumer  *wbrabbit.Consumer
}

// StorageProvider defines the repository methods needed by the consumer.
type StorageProvider interface {
	CancelBooking(ctx context.Context, bookingID, eventID string) error
	GetBookingStatus(ctx context.Context, id string) (booking.Status, error)
}

// TelegramProvider defines the interface for sending Telegram notifications.
type TelegramProvider interface {
	Send(tg, eventName string, persons int) error
}

// EmailProvider defines the interface for sending email notifications.
type EmailProvider interface {
	Send(email, eventName string, persons int) error
}

// NewBroker creates a new RabbitMQ Broker.
func NewBroker(cfg *config.AppConfig, repo StorageProvider, email EmailProvider, tg TelegramProvider) (*Broker, error) {
	rabbitDSN := fmt.Sprintf(
		"amqp://%s:%s@%s:%d/",
		cfg.RabbitMQ.User,
		cfg.RabbitMQ.Password,
		cfg.RabbitMQ.Host,
		cfg.RabbitMQ.Port,
	)

	rabbitConfig := wbrabbit.ClientConfig{
		URL:            rabbitDSN,
		ConnectionName: cfg.RabbitMQ.ConnectionName,
		ConnectTimeout: time.Duration(cfg.RabbitMQ.ConnectionTimeout) * time.Second,
		Heartbeat:      time.Duration(cfg.RabbitMQ.Heartbeat) * time.Second,
		PublishRetry: retry.Strategy{
			Attempts: cfg.Retry.Attempts,
			Delay:    cfg.Retry.Delay,
			Backoff:  cfg.Retry.Backoffs,
		},
		ConsumeRetry: retry.Strategy{
			Attempts: cfg.Retry.Attempts,
			Delay:    cfg.Retry.Delay,
			Backoff:  cfg.Retry.Backoffs,
		},
	}

	client, err := wbrabbit.NewClient(rabbitConfig)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("bad rabbitmq client connection")
		return nil, err
	}

	if err = declareInfrastructure(client, cfg); err != nil {
		wbzlog.Logger.Error().Err(err).Msg("bad rabbitmq declareInfrastructure")
		return nil, err
	}

	publisher := wbrabbit.NewPublisher(client, "booking.delay.exchange", "application/json")
	consumer := wbrabbit.NewConsumer(client, wbrabbit.ConsumerConfig{
		Queue:         "expired.queue",
		ConsumerTag:   "booking-expired-worker",
		AutoAck:       false,
		PrefetchCount: 10,
		Workers:       5,
	}, bookingExpiredHandler(repo, email, tg))

	go func() {
		if err := consumer.Start(context.Background()); err != nil {
			wbzlog.Logger.Error().Err(err).Msg("failed to start RabbitMQ consumer")
			os.Exit(1)
		}
	}()

	return &Broker{
		client:    client,
		publisher: publisher,
		consumer:  consumer,
	}, nil
}

// Close closes the RabbitMQ connection.
func (b *Broker) Close() error {
	return b.client.Close()
}

func declareInfrastructure(client *wbrabbit.RabbitClient, cfg *config.AppConfig) error {
	if err := client.DeclareExchange("booking.dlx.exchange", "direct", true, false, false, nil); err != nil {
		return err
	}

	if err := client.DeclareQueue("expired.queue", "booking.dlx.exchange", "booking.expired", true, false, true, nil); err != nil {
		return err
	}

	if err := client.DeclareExchange("booking.delay.exchange", "direct", true, false, false, nil); err != nil {
		return err
	}

	for ttlMinutes := range cfg.Event.SupportedTTLs {
		ttlMs := ttlMinutes * 60 * 1000

		queueName := fmt.Sprintf("delay_%d.queue", ttlMinutes)
		routingKey := fmt.Sprintf("delay_%d", ttlMinutes)

		args := amqp091.Table{
			"x-message-ttl":             ttlMs,
			"x-dead-letter-exchange":    "booking.dlx.exchange",
			"x-dead-letter-routing-key": "booking.expired",
		}

		if err := client.DeclareQueue(queueName, "booking.delay.exchange", routingKey, true, false, true, args); err != nil {
			return err
		}
	}

	return nil
}
