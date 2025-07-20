-- +goose Up
-- +goose StatementBegin
CREATE TABLE item_tags(
    id SERIAL PRIMARY KEY,
    item_id INT NOT NUll REFERENCES items(id) ON DELETE CASCADE,
    tag_id INT NOT NUll REFERENCES tags(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE item_tags;
-- +goose StatementEnd
