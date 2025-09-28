-- +goose Up
CREATE INDEX idx_outbox ON outbox (status, id)
WHERE status != 'sent';

-- +goose Down
DROP INDEX IF EXISTS idx_outbox;