package rabbit

import (
	"context"
	"eventbooker/internal/domain/booking"
	"encoding/json"
	wbzlog "github.com/wb-go/wbf/zlog"
)

func (s *RabbitService) PublishMsg(booking *booking.Booking) error {
	context := context.Background()

	msg, err := json.Marshal(booking)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("failed to marshal booking message")
		return err
	}
	err = s.publisher.Publish(context, msg, "booking.delay.exchange")
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("failed to publish booking message")
		return err
	}
	return nil
}