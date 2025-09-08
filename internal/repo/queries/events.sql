-- name: CreateEvent :one
INSERT INTO events (user_id, application_id, type, payload_json)
VALUES (
  sqlc.arg('user_id'),
  sqlc.arg('application_id'),
  sqlc.arg('type'),
  sqlc.arg('payload_json')
)
RETURNING id, user_id, application_id, type, payload_json, created_at;

-- name: ListEventsByApplication :many
SELECT id, user_id, application_id, type, payload_json, created_at
FROM events
WHERE application_id = sqlc.arg('application_id')
ORDER BY created_at DESC
LIMIT  sqlc.arg('limit')
OFFSET sqlc.arg('offset');
