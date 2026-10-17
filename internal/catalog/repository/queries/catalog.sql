-- name: CreateCategory :one
INSERT INTO category (name, slug, parent_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListCategories :many
SELECT *
FROM category
ORDER BY name;

-- name: CreateProduct :one
INSERT INTO product (seller_id, title, description, category_id, price_amount, price_currency)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;

-- name: GetProduct :one
SELECT *
FROM product
WHERE id = $1;

-- name: BatchGetProducts :many
SELECT *
FROM product
WHERE id = ANY (sqlc.arg('ids')::uuid[]);

-- name: ListProducts :many
SELECT *
FROM product
WHERE status = 'active'
  AND (sqlc.narg('category_id')::uuid IS NULL OR category_id = sqlc.narg('category_id'))
  AND (sqlc.narg('seller_id')::uuid IS NULL OR seller_id = sqlc.narg('seller_id'))
  AND (sqlc.narg('cursor_created_at')::timestamptz IS NULL
       OR (created_at, id) < (sqlc.narg('cursor_created_at')::timestamptz, sqlc.narg('cursor_id')::uuid))
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg('limit');

-- name: UpdateProduct :one
UPDATE product
SET title          = COALESCE(sqlc.narg('title'), title),
    description    = COALESCE(sqlc.narg('description'), description),
    price_amount   = COALESCE(sqlc.narg('price_amount'), price_amount),
    price_currency = COALESCE(sqlc.narg('price_currency'), price_currency),
    updated_at     = now()
WHERE id = sqlc.arg('id')
  AND status <> 'archived'
RETURNING *;

-- name: ArchiveProduct :one
UPDATE product
SET status     = 'archived',
    updated_at = now()
WHERE id = $1
  AND status <> 'archived'
RETURNING *;

-- name: InsertOutboxEvent :exec
INSERT INTO outbox (event_type, aggregate_id, trace_id, payload)
VALUES ($1, $2, $3, $4);

-- name: FetchUnpublishedEvents :many
SELECT *
FROM outbox
WHERE published_at IS NULL
ORDER BY created_at
LIMIT $1 FOR UPDATE SKIP LOCKED;

-- name: MarkEventsPublished :exec
UPDATE outbox
SET published_at = now()
WHERE id = ANY (sqlc.arg('ids')::uuid[]);
