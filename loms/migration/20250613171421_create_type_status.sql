-- +goose Up
CREATE TYPE order_status AS ENUM (
    'new',
    'awaiting payment', 
    'failed', 
    'payed', 
    'cancelled'
);

-- +goose Down
DROP TYPE IF EXISTS order_status;
