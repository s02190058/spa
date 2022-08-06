CREATE TABLE IF NOT EXISTS users
(
    id                 BIGSERIAL PRIMARY KEY,
    name               TEXT UNIQUE NOT NULL,
    encrypted_password TEXT        NOT NULL
);

CREATE TABLE IF NOT EXISTS types
(
    id   SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

INSERT INTO types (name)
VALUES ('link'),
       ('text')
;

CREATE TABLE categories
(
    id   SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

INSERT INTO categories (name)
VALUES ('music'),
       ('funny'),
       ('videos'),
       ('programming'),
       ('news'),
       ('fashion')
;

CREATE TABLE IF NOT EXISTS posts
(
    id          BIGSERIAL PRIMARY KEY,
    type_id     INT         NOT NULL,
    category_id INT         NOT NULL,
    title       TEXT        NOT NULL,
    text        TEXT        NOT NULL DEFAULT '',
    url         TEXT        NOT NULL DEFAULT '',
    user_id     BIGINT      NOT NULL,
    views       INT         NOT NULL DEFAULT 0,
    created     TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE posts
    ADD FOREIGN KEY (type_id) REFERENCES types (id);

ALTER TABLE posts
    ADD FOREIGN KEY (category_id) REFERENCES categories (id);

ALTER TABLE posts
    ADD FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;

CREATE TABLE IF NOT EXISTS votes
(
    post_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    vote    INT    NOT NULL
);

ALTER TABLE votes
    ADD FOREIGN KEY (post_id) REFERENCES posts (id) ON DELETE CASCADE;

ALTER TABLE votes
    ADD FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;

ALTER TABLE votes
    ADD PRIMARY KEY (post_id, user_id);

CREATE TABLE IF NOT EXISTS comments
(
    id      BIGSERIAL PRIMARY KEY,
    post_id BIGINT      NOT NULL,
    user_id BIGINT      NOT NULL,
    body    TEXT        NOt NULL,
    created TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE comments
    ADD FOREIGN KEY (post_id) REFERENCES posts (id) ON DELETE CASCADE;

CREATE INDEX ON comments (post_id);

ALTER TABLE comments
    ADD FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;
