CREATE TABLE messages (
    id           bigserial PRIMARY KEY,
    user_id      text NOT NULL,
    message_id   text NOT NULL,
    content      text NOT NULL,
    status       text NOT NULL DEFAULT 'PENDING',
    created_at   timestamptz NOT NULL DEFAULT now(),
    processed_at timestamptz,
    CONSTRAINT messages_user_message_unique UNIQUE (user_id, message_id),
    CONSTRAINT messages_status_check CHECK (status IN ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED'))
);

CREATE INDEX messages_user_id_created_at_idx ON messages (user_id, created_at DESC);
