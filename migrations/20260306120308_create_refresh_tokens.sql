-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS refresh_tokens
(
    id         uuid PRIMARY KEY     DEFAULT gen_random_uuid(),

    user_id    uuid REFERENCES users (id) ON DELETE CASCADE,

    token      text        NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL,
    revoked    boolean     NOT NULL DEFAULT FALSE,

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens (token);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS refresh_tokens;
-- +goose StatementEnd
