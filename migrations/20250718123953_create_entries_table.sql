-- +goose Up
-- +goose StatementBegin
CREATE TABLE entries(
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    item_id INT NOT NUll REFERENCES items(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE entries;
-- +goose StatementEnd
