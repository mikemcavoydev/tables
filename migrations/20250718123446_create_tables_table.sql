-- +goose Up
-- +goose StatementBegin
CREATE TABLE tables(
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE tables;
-- +goose StatementEnd
