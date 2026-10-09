-- +goose Up
CREATE TYPE order_status AS ENUM ('pending', 'reserved', 'confirmed', 'cancelled');
CREATE TYPE cancel_reason AS ENUM ('out_of_stock', 'payment_failed', 'timeout');

CREATE TABLE orders (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID        NOT NULL,
    status           order_status NOT NULL DEFAULT 'pending',
    total_amount     BIGINT      NOT NULL DEFAULT 0,
    currency         TEXT        NOT NULL DEFAULT 'RUB',
    cancel_reason    cancel_reason,
    idempotency_key  UUID        NOT NULL UNIQUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE order_item (
    id          UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id    UUID    NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    product_id  UUID    NOT NULL,
    quantity    INT     NOT NULL CHECK (quantity > 0),
    unit_price  BIGINT  NOT NULL,
    currency    TEXT    NOT NULL DEFAULT 'RUB'
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

-- +goose Down
DROP INDEX idx_outbox_unpublished;
DROP TABLE inbox;
DROP TABLE outbox;
DROP TABLE order_item;
DROP TABLE orders;
DROP TYPE cancel_reason;
DROP TYPE order_status;