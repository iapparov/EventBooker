CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY,
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date TIMESTAMP NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    total_seats INTEGER NOT NULL,
    available_seats INTEGER NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    booking_ttl INTEGER NOT NULL
);