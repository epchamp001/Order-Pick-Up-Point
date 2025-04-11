-- +goose Up
CREATE TABLE product (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    date_time TIMESTAMPTZ NOT NULL,
    type VARCHAR(50) NOT NULL,
    reception_id UUID NOT NULL REFERENCES reception(id) ON DELETE CASCADE
);
-- +goose Down
DROP TABLE IF EXISTS product;
