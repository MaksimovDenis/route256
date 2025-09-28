-- +goose Up
CREATE TABLE stocks (
    sku BIGINT PRIMARY KEY,
    total_count BIGINT NOT NULL DEFAULT 0,
    reserved BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP,
    CHECK (reserved <= total_count)
);

-- +goose Down
DROP TABLE IF EXISTS stocks;
