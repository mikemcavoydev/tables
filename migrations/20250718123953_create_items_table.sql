-- +goose Up
-- +goose StatementBegin
CREATE TABLE items(
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    table_id INT NOT NUll REFERENCES tables(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE items;
-- +goose StatementEnd
