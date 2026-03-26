package rabbit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"eventbooker/internal/domain/booking"

	wbzlog "github.com/wb-go/wbf/zlog"
)

// PublishMsg publishes a booking message to the delay exchange.
func (b *Broker) PublishMsg(ctx context.Context, bk *booking.Booking) error {
	msg, err := json.Marshal(bk)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("failed to marshal booking message")
		return err
	}

	ttlMinutes := 1
	if !bk.CreatedAt.IsZero() && !bk.ExpiredAt.IsZero() {
		diff := bk.ExpiredAt.Sub(bk.CreatedAt)
		if diff > 0 {
			ttlMinutes = int(diff.Round(time.Minute).Minutes())
			if ttlMinutes <= 0 {
				ttlMinutes = 1
			}
		}
	}

	routingKey := fmt.Sprintf("delay_%d", ttlMinutes)

	if err = b.publisher.Publish(ctx, msg, routingKey); err != nil {
		wbzlog.Logger.Error().Err(err).Msg("failed to publish booking message")
		return err
	}

	wbzlog.Logger.Info().Msgf("published booking message (routing=%s): %s", routingKey, string(msg))
	return nil
}
