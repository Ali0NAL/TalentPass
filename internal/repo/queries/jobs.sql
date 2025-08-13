-- name: CreateJob :one
INSERT INTO jobs (org_id, title, company, url, location, tags)
VALUES (sqlc.narg('org_id'), sqlc.arg('title'), sqlc.arg('company'), sqlc.narg('url'), sqlc.narg('location'), sqlc.arg('tags'))
RETURNING *;

-- name: GetJobByID :one
SELECT *
FROM jobs
WHERE id = sqlc.arg('id');

-- name: ListJobs :many
SELECT *
FROM jobs
WHERE (sqlc.narg('company')::text IS NULL OR company ILIKE '%' || sqlc.narg('company') || '%')
  AND (sqlc.narg('title')::text   IS NULL OR title   ILIKE '%' || sqlc.narg('title')   || '%')
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: UpdateJob :one
UPDATE jobs
SET
  title    = COALESCE(sqlc.narg('title'), title),
  company  = COALESCE(sqlc.narg('company'), company),
  url      = COALESCE(sqlc.narg('url'), url),
  location = COALESCE(sqlc.narg('location'), location),
  tags     = COALESCE(sqlc.narg('tags'), tags),
  updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteJob :exec
DELETE FROM jobs
WHERE id = sqlc.arg('id');
