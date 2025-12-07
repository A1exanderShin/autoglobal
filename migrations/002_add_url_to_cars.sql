-- +goose Up
ALTER TABLE cars ADD COLUMN url TEXT UNIQUE;

-- +goose Down
ALTER TABLE cars DROP COLUMN url;
