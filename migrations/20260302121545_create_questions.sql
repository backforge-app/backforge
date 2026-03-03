-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS questions
(
    id         uuid PRIMARY KEY     DEFAULT gen_random_uuid(),

    title      text        NOT NULL,
    content    jsonb       NOT NULL,
    level      smallint    NOT NULL, -- 0 = beginner, 1 = medium, 2 = advanced

    topic_id   uuid REFERENCES topics (id),

    is_free    boolean              DEFAULT FALSE,

    created_by uuid REFERENCES users (id),
    updated_by uuid REFERENCES users (id),

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_questions_topic_id ON questions (topic_id);

CREATE INDEX idx_questions_level ON questions (level);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS questions;
-- +goose StatementEnd
