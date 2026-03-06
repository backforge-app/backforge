-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_topic_progress
(
    id               uuid PRIMARY KEY     DEFAULT gen_random_uuid(),

    user_id          uuid REFERENCES users (id) ON DELETE CASCADE,
    topic_id         uuid REFERENCES topics (id) ON DELETE CASCADE,

    current_position int                  DEFAULT 0,

    updated_at       timestamptz NOT NULL DEFAULT now(),

    UNIQUE (user_id, topic_id)
);

CREATE INDEX idx_utp_user_id ON user_topic_progress (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_topic_progress;
-- +goose StatementEnd
