-- +goose Up
CREATE TABLE cars (
                      id SERIAL PRIMARY KEY,
                      brand TEXT NOT NULL,
                      model TEXT NOT NULL,
                      year INT NOT NULL CHECK (year >= 1900 AND year <= EXTRACT(YEAR FROM NOW())),
    price INT NOT NULL CHECK (price >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE cars;

-- goose -dir ./migrations postgres "postgres://postgres:postgres@localhost:5432/autoglobal?sslmode=disable" up