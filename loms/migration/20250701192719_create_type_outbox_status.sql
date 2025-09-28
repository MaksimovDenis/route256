-- +goose Up
CREATE TYPE outbox_status AS ENUM (
    'new',
    'sent', 
    'pending',
    'error'
);

-- +goose Down
DROP TYPE IF EXISTS outbox_status;
