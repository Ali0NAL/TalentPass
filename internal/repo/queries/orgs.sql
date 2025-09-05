-- name: CreateOrganization :one
INSERT INTO organizations (name) VALUES ($1)
RETURNING *;

-- name: AddOrgMember :exec
INSERT INTO org_members (org_id, user_id, role) VALUES ($1, $2, $3)
ON CONFLICT (org_id, user_id) DO UPDATE SET role = EXCLUDED.role;

-- name: ListMyOrganizations :many
SELECT o.*
FROM organizations o
JOIN org_members m ON m.org_id = o.id
WHERE m.user_id = $1
ORDER BY o.created_at DESC;

-- name: GetOrgMemberRole :one
SELECT role
FROM org_members
WHERE org_id = $1 AND user_id = $2;

-- name: GetOrganization :one
SELECT * FROM organizations WHERE id = $1;
-- name: DeleteOrganization :exec