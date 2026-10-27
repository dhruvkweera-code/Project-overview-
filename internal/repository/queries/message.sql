-- name: CreateMessage :one
INSERT INTO messages (user_id, message_id, content)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, message_id) DO NOTHING
RETURNING *;

-- name: GetMessageByUserAndMessageID :one
SELECT * FROM messages WHERE user_id = $1 AND message_id = $2;

-- name: UpdateMessageStatus :one
UPDATE messages
SET status = $3, processed_at = $4
WHERE user_id = $1 AND message_id = $2
RETURNING *;
