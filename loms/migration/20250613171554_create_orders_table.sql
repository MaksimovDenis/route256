-- +goose Up
CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    status order_status NOT NULL DEFAULT 'new',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS orders;
