-- name: CreateEvent :exec
INSERT INTO outbox (topic, key, payload)
VALUES ($1, $2, $3);

-- name: FetchNextMessages :many
SELECT id, topic, key, payload, status 
FROM outbox
WHERE status = 'new'
ORDER BY id
LIMIT $1
FOR UPDATE SKIP LOCKED;

-- name: MarkAsSent :exec
UPDATE outbox
SET
    status = 'sent',
    sent_at = NOW()
WHERE id = ANY(sqlc.arg(ids)::bigint[]);

-- name: MarkAsError :exec
UPDATE outbox
SET
    status = 'error',
    sent_at = NOW()
WHERE id = ANY(sqlc.arg(ids)::bigint[]);