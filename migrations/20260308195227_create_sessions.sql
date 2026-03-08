-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS sessions
(
    id         uuid PRIMARY KEY     DEFAULT gen_random_uuid(),

    user_id    uuid REFERENCES users (id) ON DELETE CASCADE,

    token      text        NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL,
    revoked    boolean     NOT NULL DEFAULT FALSE,

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_sessions_user_id ON sessions (user_id);
CREATE INDEX idx_sessions_token ON sessions (token);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sessions;
-- +goose StatementEnd
