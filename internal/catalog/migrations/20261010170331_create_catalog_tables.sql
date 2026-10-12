-- +goose Up
CREATE TYPE product_status AS ENUM ('draft', 'active', 'archived');

CREATE TABLE category (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL,
    slug       TEXT NOT NULL UNIQUE,
    parent_id  UUID REFERENCES category (id)
);

CREATE TABLE product (
    id              UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    seller_id       UUID           NOT NULL,
    title           TEXT           NOT NULL CHECK (length(title) BETWEEN 1 AND 255),
    description     TEXT           NOT NULL DEFAULT '' CHECK (length(description) <= 5000),
    category_id     UUID           NOT NULL REFERENCES category (id),
    price_amount    BIGINT         NOT NULL CHECK (price_amount >= 0),
    price_currency  CHAR(3)        NOT NULL DEFAULT 'RUB' CHECK (price_currency ~ '^[A-Z]{3}$'),
    status          product_status NOT NULL DEFAULT 'draft',
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ    NOT NULL DEFAULT now()
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

CREATE INDEX idx_category_parent_id ON category (parent_id);
CREATE INDEX idx_product_seller_id ON product (seller_id);
CREATE INDEX idx_product_category_id ON product (category_id);
CREATE INDEX idx_product_created_at_id ON product (created_at DESC, id DESC);
CREATE INDEX idx_outbox_unpublished ON outbox (created_at) WHERE published_at IS NULL;

-- +goose Down
DROP TABLE outbox;
DROP TABLE product;
DROP TABLE category;
DROP TYPE product_status;
