-- name: CreateOrder :one
INSERT INTO orders (user_id)
VALUES ($1)
RETURNING id;

-- name: GetByOrderID :many
SELECT 
    o.user_id,
    o.status,
    oi.sku,
    oi.count
FROM orders o
JOIN order_items oi ON o.id = oi.order_id
WHERE o.id = $1;

-- name: SetStatus :exec
UPDATE orders
SET status = $1,
    updated_at = NOW()
WHERE id = $2;

-- name: GetByOrderIDForUpdate :many
SELECT 
    o.user_id,
    o.status,
    oi.sku,
    oi.count
FROM orders o
JOIN order_items oi ON o.id = oi.order_id
WHERE o.id = $1
FOR UPDATE;

-- name: CreateOrderItems :exec
INSERT INTO order_items (order_id, sku, count)
SELECT unnest(@order_ids::bigint[]), unnest(@skus::bigint[]), unnest(@counts::bigint[]);
