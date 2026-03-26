package rabbit

import (
	"context"
	"encoding/json"
	"fmt"

	"eventbooker/internal/domain/booking"

	"github.com/rabbitmq/amqp091-go"
	wbrabbit "github.com/wb-go/wbf/rabbitmq"
	wbzlog "github.com/wb-go/wbf/zlog"
)

func bookingExpiredHandler(repo StorageProvider, email EmailProvider, tg TelegramProvider) wbrabbit.MessageHandler {
	return func(ctx context.Context, msg amqp091.Delivery) error {
		wbzlog.Logger.Info().Msgf("received booking expired message: %s", string(msg.Body))

		var payload booking.Booking
		if err := json.Unmarshal(msg.Body, &payload); err != nil {
			wbzlog.Logger.Error().Err(err).Msg("failed to unmarshal booking expired message")
			return fmt.Errorf("invalid payload: %w", err)
		}

		status, err := repo.GetBookingStatus(ctx, payload.ID.String())
		if err != nil {
			return fmt.Errorf("get booking status failed: %w", err)
		}

		if status == booking.StatusCancelled || status == booking.StatusConfirmed {
			wbzlog.Logger.Info().Msgf("booking %s already processed with status %s, skipping", payload.ID.String(), status)
			return nil
		}

		if err := repo.CancelBooking(ctx, payload.ID.String(), payload.EventID.String()); err != nil {
			return err
		}

		if payload.EmailNotification {
			if err := email.Send(payload.EmailRecepient, payload.EventName, payload.Count); err != nil {
				wbzlog.Logger.Error().Err(err).Msg("cannot send email notification")
			}
		}

		if payload.TelegramNotification {
			if err := tg.Send(payload.TelegramRecepient, payload.EventName, payload.Count); err != nil {
				wbzlog.Logger.Error().Err(err).Msg("cannot send telegram notification")
			}
		}

		return nil
	}
}
