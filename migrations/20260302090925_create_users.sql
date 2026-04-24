-- +goose Up
-- +goose StatementBegin
CREATE TYPE user_role AS ENUM ('user', 'admin');

CREATE TABLE IF NOT EXISTS users
(
    id                uuid PRIMARY KEY             DEFAULT gen_random_uuid(),

    email             varchar(255) UNIQUE NOT NULL,
    password_hash     text,

    username          varchar(32) UNIQUE,
    first_name        varchar(64),
    last_name         varchar(64),
    photo_url         text,

    role              user_role           NOT NULL DEFAULT 'user',
    is_email_verified boolean             NOT NULL DEFAULT false,

    created_at        timestamptz         NOT NULL DEFAULT now(),
    updated_at        timestamptz         NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS user_role;
-- +goose StatementEnd
