-- +goose Up
-- +goose StatementBegin
CREATE TYPE progress_status AS ENUM ('new', 'skipped', 'known', 'learned');

CREATE TABLE IF NOT EXISTS user_question_progress
(
    id          uuid PRIMARY KEY         DEFAULT gen_random_uuid(),

    user_id     uuid REFERENCES users (id) ON DELETE CASCADE,
    question_id uuid REFERENCES questions (id) ON DELETE CASCADE,

    status      progress_status NOT NULL DEFAULT 'new',

    updated_at  timestamptz     NOT NULL DEFAULT now(),

    UNIQUE (user_id, question_id)
);

CREATE INDEX idx_user_question_progress_user
    ON user_question_progress (user_id);

CREATE INDEX idx_user_question_progress_question
    ON user_question_progress (question_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_question_progress;

DROP TYPE IF EXISTS progress_status;
-- +goose StatementEnd
