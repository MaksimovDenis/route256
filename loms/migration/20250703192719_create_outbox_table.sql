-- +goose Up
CREATE TABLE outbox (
    id BIGSERIAL PRIMARY KEY,
    topic TEXT NOT NULL,
    key TEXT,
    payload JSONB NOT NULL,
    status outbox_status NOT NULL DEFAULT 'new',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    sent_at TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS outbox;
