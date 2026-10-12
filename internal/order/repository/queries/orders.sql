-- name: CreateOrder :one
INSERT INTO orders (user_id, total_amount, idempotency_key)
VALUES ($1, $2, $3)
ON CONFLICT (idempotency_key) DO NOTHING
RETURNING id, status;

-- name: GetOrderByIdempotencyKey :one
SELECT id, status
FROM orders
WHERE idempotency_key = $1;

-- name: CreateOrderItems :copyfrom
INSERT INTO order_item (order_id, product_id, quantity, unit_price)
VALUES ($1, $2, $3, $4);

-- name: InsertOutboxEvent :exec
INSERT INTO outbox (event_type, aggregate_id, trace_id, payload)
VALUES ($1, $2, $3, $4);