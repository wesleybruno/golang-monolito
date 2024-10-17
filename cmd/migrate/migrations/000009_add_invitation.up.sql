CREATE TABLE IF NOT EXISTS user_invitation(
    token bytea PRIMARY KEY,
    user_id bigint NOT NULL
);
