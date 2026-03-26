package postgres

import (
	"context"
	"fmt"

	"eventbooker/internal/domain/booking"
	"eventbooker/internal/domain/event"

	"github.com/wb-go/wbf/retry"
	wbzlog "github.com/wb-go/wbf/zlog"
)

// CreateBooking inserts a new booking and decrements available seats atomically.
func (r *Repository) CreateBooking(ctx context.Context, b *booking.Booking) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	tx, err := r.db.Master.BeginTx(ctx, nil)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("cannot start transaction in create_booking")
		return err
	}
	defer func() { _ = tx.Rollback() }()

	insertQuery := `
		INSERT INTO bookings (id, event_id, user_id, count, price, status, created_at, expired_at,
			telegram_notification, email_notification, telegram_recepient, email_recepient)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	err = retry.DoContext(ctx, retry.Strategy{Attempts: r.retry.Attempts, Delay: r.retry.Delay, Backoff: r.retry.Backoffs}, func() error {
		_, err := tx.ExecContext(ctx, insertQuery,
			b.ID, b.EventID, b.UserID, b.Count, b.Price, b.Status,
			b.CreatedAt, b.ExpiredAt, b.TelegramNotification, b.EmailNotification,
			b.TelegramRecepient, b.EmailRecepient,
		)
		return err
	})
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("failed to insert booking")
		return err
	}

	updateQuery := `
		UPDATE events
		SET available_seats = available_seats - $1
		WHERE id = $2 AND available_seats >= $1
	`

	err = retry.DoContext(ctx, retry.Strategy{Attempts: r.retry.Attempts, Delay: r.retry.Delay, Backoff: r.retry.Backoffs}, func() error {
		result, err := tx.ExecContext(ctx, updateQuery, b.Count, b.EventID)
		if err != nil {
			return err
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if rows == 0 {
			return fmt.Errorf("not enough available seats for the event")
		}

		return nil
	})
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		wbzlog.Logger.Error().Err(err).Msg("failed to commit transaction")
		return err
	}

	return nil
}

// ConfirmBooking sets a booking's status to confirmed.
func (r *Repository) ConfirmBooking(ctx context.Context, id string) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	query := `UPDATE bookings SET status = 'confirmed' WHERE id = $1`
	_, err := r.db.ExecWithRetry(ctx, retry.Strategy{Attempts: r.retry.Attempts, Delay: r.retry.Delay, Backoff: r.retry.Backoffs}, query, id)
	return err
}

// CancelBooking cancels a booking and returns the seats to the event.
func (r *Repository) CancelBooking(ctx context.Context, bookingID, eventID string) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	tx, err := r.db.Master.BeginTx(ctx, nil)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("cannot start transaction in cancel_booking")
		return err
	}
	defer func() { _ = tx.Rollback() }()

	cancelQuery := `UPDATE bookings SET status = 'cancelled' WHERE id = $1`
	err = retry.DoContext(ctx, retry.Strategy{Attempts: r.retry.Attempts, Delay: r.retry.Delay, Backoff: r.retry.Backoffs}, func() error {
		result, err := tx.ExecContext(ctx, cancelQuery, bookingID)
		if err != nil {
			return err
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if rows == 0 {
			return fmt.Errorf("booking already cancelled")
		}

		return nil
	})
	if err != nil {
		return err
	}

	returnSeatsQuery := `
		UPDATE events
		SET available_seats = available_seats + (SELECT count FROM bookings WHERE id = $1)
		WHERE id = $2
	`

	err = retry.DoContext(ctx, retry.Strategy{Attempts: r.retry.Attempts, Delay: r.retry.Delay, Backoff: r.retry.Backoffs}, func() error {
		result, err := tx.ExecContext(ctx, returnSeatsQuery, bookingID, eventID)
		if err != nil {
			return err
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if rows == 0 {
			return fmt.Errorf("failed to increase available seats")
		}

		return nil
	})
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		wbzlog.Logger.Error().Err(err).Msg("failed to commit transaction")
		return err
	}

	return nil
}

// GetBookingStatus returns the status of a booking.
func (r *Repository) GetBookingStatus(ctx context.Context, id string) (booking.Status, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	query := `SELECT status FROM bookings WHERE id = $1`

	row, err := r.db.QueryRowWithRetry(ctx, retry.Strategy{Attempts: r.retry.Attempts, Delay: r.retry.Delay, Backoff: r.retry.Backoffs}, query, id)
	if err != nil {
		return "", err
	}

	var status booking.Status
	if err = row.Scan(&status); err != nil {
		return "", err
	}

	return status, nil
}

// CreateEvent inserts a new event.
func (r *Repository) CreateEvent(ctx context.Context, e *event.Event) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	query := `
		INSERT INTO events (id, creator_id, date, name, description, total_seats, available_seats, price, booking_ttl)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecWithRetry(ctx, retry.Strategy{Attempts: r.retry.Attempts, Delay: r.retry.Delay, Backoff: r.retry.Backoffs}, query,
		e.ID, e.CreatorID, e.Date, e.Name, e.Description, e.MaxCountPeople, e.FreePlaces, e.Price, e.BookingTTL,
	)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("failed to insert event")
		return err
	}

	return nil
}

// GetEvent retrieves an event by ID, including its bookings.
func (r *Repository) GetEvent(ctx context.Context, eventID string) (*event.Event, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	query := `
		SELECT id, creator_id, date, name, description, total_seats, available_seats, price, booking_ttl
		FROM events WHERE id = $1
	`

	row, err := r.db.QueryRowWithRetry(ctx, retry.Strategy{Attempts: r.retry.Attempts, Delay: r.retry.Delay, Backoff: r.retry.Backoffs}, query, eventID)
	if err != nil {
		return nil, err
	}

	var ev event.Event
	if err = row.Scan(
		&ev.ID, &ev.CreatorID, &ev.Date, &ev.Name, &ev.Description,
		&ev.MaxCountPeople, &ev.FreePlaces, &ev.Price, &ev.BookingTTL,
	); err != nil {
		return nil, err
	}

	bookingsQuery := `
		SELECT id, event_id, user_id, count, price, status, created_at, expired_at,
			telegram_notification, email_notification, telegram_recepient, email_recepient
		FROM bookings WHERE event_id = $1
	`

	rows, err := r.db.QueryWithRetry(ctx, retry.Strategy{Attempts: r.retry.Attempts, Delay: r.retry.Delay, Backoff: r.retry.Backoffs}, bookingsQuery, eventID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var bookings []*booking.Booking
	for rows.Next() {
		var b booking.Booking
		if err = rows.Scan(
			&b.ID, &b.EventID, &b.UserID, &b.Count, &b.Price, &b.Status,
			&b.CreatedAt, &b.ExpiredAt, &b.TelegramNotification, &b.EmailNotification,
			&b.TelegramRecepient, &b.EmailRecepient,
		); err != nil {
			return nil, err
		}
		b.EventName = ev.Name
		bookings = append(bookings, &b)
	}

	ev.Bookings = bookings
	return &ev, nil
}
