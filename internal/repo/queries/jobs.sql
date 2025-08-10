-- name: CreateJob :one
INSERT INTO jobs (org_id, title, company, url, location, tags)
VALUES (sqlc.narg('org_id'), sqlc.arg('title'), sqlc.arg('company'), sqlc.narg('url'), sqlc.narg('location'), sqlc.arg('tags'))
RETURNING *;

-- name: ListJobs :many
SELECT *
FROM jobs
WHERE (sqlc.narg('org_id')::bigint IS NULL OR org_id = sqlc.narg('org_id'))
  AND (sqlc.narg('company')::text IS NULL OR company ILIKE '%' || sqlc.narg('company') || '%')
  AND (sqlc.narg('title')::text   IS NULL OR title   ILIKE '%' || sqlc.narg('title')   || '%')
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');
