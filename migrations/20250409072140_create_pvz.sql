-- +goose Up
CREATE TABLE pvz (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    registration_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    city VARCHAR(50) NOT NULL
);
-- +goose Down
DROP TABLE IF EXISTS pvz;

