-- +goose Up
ALTER TABLE cars
    ADD COLUMN user_id BIGINT REFERENCES users(id);

CREATE INDEX idx_cars_user_id ON cars(user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_cars_user_id;
ALTER TABLE cars DROP COLUMN user_id;
