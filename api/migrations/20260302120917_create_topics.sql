-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS topics
(
    id          uuid PRIMARY KEY     DEFAULT gen_random_uuid(),

    title       text        NOT NULL,
    description text,

    created_by  uuid REFERENCES users (id),
    updated_by  uuid REFERENCES users (id),

    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS topics;
-- +goose StatementEnd
