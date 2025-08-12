-- name: CreateEvent :one
INSERT INTO events (user_id, application_id, type, payload_json)
VALUES (sqlc.narg('user_id'), sqlc.narg('application_id'), sqlc.arg('type'), sqlc.arg('payload_json'))
RETURNING *;
