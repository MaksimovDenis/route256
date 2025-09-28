-- name: GetStockBySku :one
SELECT sku, total_count, reserved
FROM stocks
WHERE sku = $1;

-- name: GetStocksBySkuForUpdate :many
SELECT sku, total_count, reserved 
FROM stocks 
WHERE sku = ANY(sqlc.arg(sku)::bigint[]) FOR UPDATE;

-- name: UpdateStocks :exec
UPDATE stocks s
SET
    total_count = u.total_count,
    reserved = u.reserved,
    updated_at = now()
FROM (
    SELECT 
        unnest(sqlc.arg(sku)::bigint[]) AS sku,
        unnest(sqlc.arg(total_count)::bigint[]) AS total_count,
        unnest(sqlc.arg(reserved)::bigint[]) AS reserved
) AS u
WHERE s.sku = u.sku;