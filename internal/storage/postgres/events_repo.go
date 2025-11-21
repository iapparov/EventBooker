package postgres

import (
	"context"
	"eventbooker/internal/domain/booking"
	"eventbooker/internal/domain/event"
	"fmt"
	"github.com/wb-go/wbf/retry"
	wbzlog "github.com/wb-go/wbf/zlog"
)

func (p *Postgres) CreateBooking(booking *booking.Booking) error {
	ctx := context.Background()
	tx, err := p.db.Master.BeginTx(ctx, nil)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("cant start transaction in create_booking")
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	query := `
		INSERT INTO bookings (id, event_id, user_id, count, price, status, created_at, expired_at, telegram_notification, email_notification, telegram_recepient, email_recepient)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	err = retry.DoContext(ctx, retry.Strategy{Attempts: p.cfg.Attempts, Delay: p.cfg.Delay, Backoff: p.cfg.Backoffs}, func() error {
		_, err := tx.ExecContext(ctx, query,
			booking.ID,
			booking.EventID,
			booking.UserID,
			booking.Count,
			booking.Price,
			booking.Status,
			booking.CreatedAt,
			booking.ExpiredAt,
			booking.TelegramNotification,
			booking.EmailNotification,
			booking.TelegramRecepient,
			booking.EmailRecepient,
		)
		return err
	})

	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("Failed to execute insert notification query")
		return err
	}

	queryUpdateEvent := `
		UPDATE events
		SET available_seats = available_seats - $1
		WHERE id = $2 AND available_seats >= $1
	`

	err = retry.DoContext(ctx, retry.Strategy{Attempts: p.cfg.Attempts, Delay: p.cfg.Delay, Backoff: p.cfg.Backoffs}, func() error {
		result, err := tx.ExecContext(ctx, queryUpdateEvent, booking.Count, booking.EventID)
		if err != nil {
			wbzlog.Logger.Error().Err(err).Msg("Failed to execute update event query")
			return err
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			wbzlog.Logger.Error().Err(err).Msg("Failed to get rows affected for update event query")
			return err
		}

		if rowsAffected == 0 {
			wbzlog.Logger.Error().Msg("Not enough available seats for the event")
			return fmt.Errorf("not enough available seats for the event")
		}
		return nil
	})

	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("Failed to commit transaction")
		return err
	}
	return nil
}

func (p *Postgres) ConfirmBooking(id string) error {
	ctx := context.Background()

	query := `
		UPDATE bookings
		SET status = 'confirmed'
		WHERE id = $1
	`

	_, err := p.db.ExecWithRetry(ctx, retry.Strategy{Attempts: p.cfg.Attempts, Delay: p.cfg.Delay, Backoff: p.cfg.Backoffs}, query, id)
	return err
}

func (p *Postgres) CancelBooking(ctx context.Context, bookingid, eventid string) error {

	tx, err := p.db.Master.BeginTx(ctx, nil)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("cant start transaction in cancel_booking")
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	querybooking := `
		UPDATE bookings
		SET status = 'cancelled'
		WHERE id = $1
	`

	err = retry.DoContext(ctx, retry.Strategy{Attempts: p.cfg.Attempts, Delay: p.cfg.Delay, Backoff: p.cfg.Backoffs}, func() error {
		result, err := tx.ExecContext(ctx, querybooking, bookingid)
		if err != nil {
			wbzlog.Logger.Error().Err(err).Msg("Failed to execute update booking query")
			return err
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			wbzlog.Logger.Error().Err(err).Msg("Failed to get rows affected for update booking query")
			return err
		}

		if rowsAffected == 0 {
			wbzlog.Logger.Error().Msg("booking already cancelled")
			return fmt.Errorf("booking already cancelled")
		}
		return nil
	})

	if err != nil {
		return err
	}

	queryevent := `
		UPDATE events
		SET available_seats = available_seats + (SELECT count FROM bookings WHERE id = $1)
		WHERE id = $2
	`

	err = retry.DoContext(ctx, retry.Strategy{Attempts: p.cfg.Attempts, Delay: p.cfg.Delay, Backoff: p.cfg.Backoffs}, func() error {
		result, err := tx.ExecContext(ctx, queryevent, bookingid, eventid)
		if err != nil {
			wbzlog.Logger.Error().Err(err).Msg("Failed to execute update event query")
			return err
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			wbzlog.Logger.Error().Err(err).Msg("Failed to get rows affected for update event query")
			return err
		}

		if rowsAffected == 0 {
			wbzlog.Logger.Error().Msg("error increase available seats in event")
			return fmt.Errorf("error increase available seats in event")
		}
		return nil
	})

	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("Failed to commit transaction")
		return err
	}
	return err
}

func (p *Postgres) GetBookingStatus(ctx context.Context, id string) (booking.BookingStatus, error) {
	var status booking.BookingStatus

	query := `
		SELECT status
		FROM bookings
		WHERE id = $1
	`

	row, err := p.db.QueryRowWithRetry(ctx, retry.Strategy{Attempts: p.cfg.Attempts, Delay: p.cfg.Delay, Backoff: p.cfg.Backoffs}, query, id)
	if err != nil {
		return "", err
	}

	err = row.Scan(&status)
	if err != nil {
		return "", err
	}

	return status, nil
}
func (p *Postgres) CreateEvent(event *event.Event) error {
	ctx := context.Background()

	query := `
		INSERT INTO events (id, creator_id, date, name, description, total_seats, available_seats, price, booking_ttl)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := p.db.ExecWithRetry(ctx, retry.Strategy{Attempts: p.cfg.Attempts, Delay: p.cfg.Delay, Backoff: p.cfg.Backoffs}, query,
		event.Id,
		event.CreatorId,
		event.Date,
		event.Name,
		event.Description,
		event.MaxCountPeople,
		event.FreePlaces,
		event.Price,
		event.BookingTTL,
	)

	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("Failed to execute insert event query")
		return err
	}

	return nil
}

func (p *Postgres) GetEvent(eventid string) (*event.Event, error) {

	ctx := context.Background()

	query := `
		SELECT id, creator_id, date, name, description, total_seats, available_seats, price, booking_ttl
		FROM events
		WHERE id = $1
	`
	row, err := p.db.QueryRowWithRetry(ctx, retry.Strategy{Attempts: p.cfg.Attempts, Delay: p.cfg.Delay, Backoff: p.cfg.Backoffs}, query, eventid)
	if err != nil {
		return nil, err
	}

	var ev event.Event
	err = row.Scan(
		&ev.Id,
		&ev.CreatorId,
		&ev.Date,
		&ev.Name,
		&ev.Description,
		&ev.MaxCountPeople,
		&ev.FreePlaces,
		&ev.Price,
		&ev.BookingTTL,
	)
	if err != nil {
		return nil, err
	}
	var bookings []*booking.Booking

	queryBookings := `
		SELECT id, event_id, user_id, count, price, status, created_at, expired_at, telegram_notification, email_notification, telegram_recepient, email_recepient
		FROM bookings
		WHERE event_id = $1
	`
	rows, err := p.db.QueryWithRetry(ctx, retry.Strategy{Attempts: p.cfg.Attempts, Delay: p.cfg.Delay, Backoff: p.cfg.Backoffs}, queryBookings, eventid)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var b booking.Booking
		err := rows.Scan(
			&b.ID,
			&b.EventID,
			&b.UserID,
			&b.Count,
			&b.Price,
			&b.Status,
			&b.CreatedAt,
			&b.ExpiredAt,
			&b.TelegramNotification,
			&b.EmailNotification,
			&b.TelegramRecepient,
			&b.EmailRecepient,
		)
		if err != nil {
			return nil, err
		}
		b.EventName = ev.Name
		bookings = append(bookings, &b)
	}
	ev.Bookings = bookings

	return &ev, nil
}
