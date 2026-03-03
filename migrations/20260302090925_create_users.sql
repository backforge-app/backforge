-- +goose Up
-- +goose StatementBegin
CREATE TYPE user_role AS ENUM ('user', 'admin');

CREATE TABLE IF NOT EXISTS users
(
    id            uuid PRIMARY KEY       DEFAULT gen_random_uuid(),

    tg_user_id    bigint UNIQUE NOT NULL,
    tg_username   varchar(32),
    tg_first_name varchar(64),
    tg_last_name  varchar(64),

    role          user_role     NOT NULL DEFAULT 'user',

    is_premium    boolean                DEFAULT FALSE,

    created_at    timestamptz   NOT NULL DEFAULT now(),
    updated_at    timestamptz   NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS user_role;
-- +goose StatementEnd
