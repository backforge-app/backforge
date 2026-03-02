-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS question_tags
(
    question_id uuid REFERENCES questions (id) ON DELETE CASCADE,
    tag_id      uuid REFERENCES tags (id) ON DELETE CASCADE,

    PRIMARY KEY (question_id, tag_id)
);

CREATE INDEX idx_question_tag ON question_tags (tag_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS question_tags;
-- +goose StatementEnd
