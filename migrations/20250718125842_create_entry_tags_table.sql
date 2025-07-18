-- +goose Up
-- +goose StatementBegin
CREATE TABLE entry_tags(
    id SERIAL PRIMARY KEY,
    entry_id INT NOT NUll REFERENCES items(id) ON DELETE CASCADE,
    tag_id INT NOT NUll REFERENCES tags(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE entry_tags;
-- +goose StatementEnd
