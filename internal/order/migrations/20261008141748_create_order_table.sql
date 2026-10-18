-- +goose Up
CREATE TYPE order_status AS ENUM ('pending', 'reserved', 'confirmed', 'cancelled');
CREATE TYPE cancel_reason AS ENUM ('out_of_stock', 'payment_failed', 'timeout');

CREATE TABLE orders (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID        NOT NULL,
    status           order_status NOT NULL DEFAULT 'pending',
    total_amount     BIGINT      NOT NULL CHECK (total_amount >= 0),
    currency        CHAR(3)        NOT NULL DEFAULT 'RUB' CHECK (currency ~ '^[A-Z]{3}$'),
    cancel_reason    cancel_reason,
    idempotency_key  UUID        NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT orders_user_id_idempotency_key_uniq UNIQUE (user_id, idempotency_key)
);

CREATE TABLE order_item (
    id          UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id    UUID    NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    product_id  UUID    NOT NULL,
    quantity    INT     NOT NULL CHECK (quantity > 0),
    unit_price  BIGINT  NOT NULL CHECK (unit_price >= 0),
    currency    CHAR(3)        NOT NULL DEFAULT 'RUB' CHECK (currency ~ '^[A-Z]{3}$')
);

CREATE TABLE outbox (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type   TEXT        NOT NULL,
    aggregate_id UUID        NOT NULL,
    trace_id     UUID,
    payload      BYTEA       NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at TIMESTAMPTZ
);

CREATE TABLE inbox (
    event_id     UUID        PRIMARY KEY,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_outbox_unpublished ON outbox (created_at) WHERE published_at IS NULL;
CREATE INDEX idx_order_item_order_id ON order_item (order_id);

-- +goose Down
DROP INDEX idx_order_item_order_id;
DROP INDEX idx_outbox_unpublished;
DROP TABLE inbox;
DROP TABLE outbox;
DROP TABLE order_item;
DROP TABLE orders;
DROP TYPE cancel_reason;
DROP TYPE order_status;