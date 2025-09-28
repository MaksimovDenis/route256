-- +goose Up
CREATE TABLE order_items (
    order_id BIGINT NOT NULL,
    sku BIGINT NOT NULL,
    count BIGINT NOT NULL CHECK (count > 0)
);

-- +goose Down
DROP TABLE IF EXISTS order_items;
