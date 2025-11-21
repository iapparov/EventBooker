CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    count INTEGER NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    expired_at TIMESTAMP,
    telegram_notification BOOLEAN NOT NULL,
    email_notification BOOLEAN NOT NULL,
    telegram_recepient TEXT,
    email_recepient TEXT
);