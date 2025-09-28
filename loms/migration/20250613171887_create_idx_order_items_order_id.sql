-- +goose Up
CREATE INDEX idx_order_items_order_id ON order_items(order_id);

-- +goose Down
DROP INDEX IF EXISTS idx_order_items_order_id;
