package rabbit

import (
    "github.com/rabbitmq/amqp091-go"
    wbrabbit "github.com/wb-go/wbf/rabbitmq"
    "eventbooker/internal/domain/booking"
    "context"
    "encoding/json"
    "fmt"
	wbzlog "github.com/wb-go/wbf/zlog"
)

func BookingExpiredHandler(repo StorageProvider, email EmailProvider, tg TelegramProvider) wbrabbit.MessageHandler {
    return func(ctx context.Context, msg amqp091.Delivery) error {

        var payload booking.Booking
        if err := json.Unmarshal(msg.Body, &payload); err != nil {
            wbzlog.Logger.Error().Err(err).Msg("failed to unmarshal booking expired message")
            return fmt.Errorf("invalid payload: %w", err)
        }

        status, err := repo.GetBookingStatus(ctx, payload.ID.String())
        if err != nil {
            return fmt.Errorf("get booking status failed: %w", err)
        }

        if status == booking.BookingStatusCancelled || status == booking.BookingStatusConfirmed {
            wbzlog.Logger.Info().Msgf("booking %s already processed with status %s, skipping cancellation", payload.ID.String(), status)
            return nil
        }

        if err := repo.CancelBooking(ctx, payload.ID.String(), payload.EventID.String()); err != nil {
            return err
        }

        if payload.EmailNotification {
            err = email.Send(payload.EmailRecepient, payload.EventName, payload.Count)
            if err != nil {
                return err
            }
        }

        if payload.TelegramNotification {
            err = tg.Send(payload.TelegramRecepient, payload.EventName, payload.Count)
            if err != nil {
                return err
            }
        }

        return nil
    }
}