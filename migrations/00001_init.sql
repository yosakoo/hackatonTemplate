-- +goose Up
-- +goose StatementBegin

-- Place your initial schema here.
-- Example:
-- CREATE TABLE users (
--     id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
--     name TEXT NOT NULL,
--     created_at TIMESTAMPTZ NOT NULL DEFAULT now()
-- );

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- DROP TABLE IF EXISTS users;

-- +goose StatementEnd
