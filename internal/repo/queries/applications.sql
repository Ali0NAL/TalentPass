-- name: CreateApplication :one
INSERT INTO applications (job_id, user_id, status, notes, next_action_at)
VALUES (sqlc.arg('job_id'), sqlc.arg('user_id'), sqlc.narg('status'), sqlc.narg('notes'), sqlc.narg('next_action_at'))
RETURNING *;

-- name: ListApplicationsByUser :many
SELECT *
FROM applications
WHERE user_id = sqlc.arg('user_id')
  AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status'))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: UpdateApplicationStatus :one
UPDATE applications
SET status = sqlc.arg('status'), updated_at = now()
WHERE id = sqlc.arg('id') AND user_id = sqlc.arg('user_id')
RETURNING *;
