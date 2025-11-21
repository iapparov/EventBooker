package rabbit

import (
    "context"
    "eventbooker/internal/domain/booking"
    "encoding/json"
    "fmt"
    "time"
    wbzlog "github.com/wb-go/wbf/zlog"
)

func (s *RabbitService) PublishMsg(booking *booking.Booking) error {
    ctx := context.Background()

    msg, err := json.Marshal(booking)
    if err != nil {
        wbzlog.Logger.Error().Err(err).Msg("failed to marshal booking message")
        return err
    }

    // вычисляем TTL в минутах на основе created_at / expired_at и формируем routing key для delay очереди
    ttlMinutes := 1
    if !booking.CreatedAt.IsZero() && !booking.ExpiredAt.IsZero() {
        diff := booking.ExpiredAt.Sub(booking.CreatedAt)
        if diff > 0 {
            ttlMinutes = int(diff.Round(time.Minute).Minutes())
            if ttlMinutes <= 0 {
                ttlMinutes = 1
            }
        }
    }
    routingKey := fmt.Sprintf("delay_%d", ttlMinutes)

    // публикуем в exchange с routing key соответствующим delay очереди
    if err = s.publisher.Publish(ctx, msg, routingKey); err != nil {
        wbzlog.Logger.Error().Err(err).Msg("failed to publish booking message")
        return err
    }

    wbzlog.Logger.Info().Msgf("published booking message (routing=%s): %s", routingKey, string(msg))
    return nil
}