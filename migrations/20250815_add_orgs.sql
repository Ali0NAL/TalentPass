-- +goose Up
CREATE TABLE organizations (
  id          BIGSERIAL PRIMARY KEY,
  name        TEXT NOT NULL UNIQUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE org_members (
  org_id     BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  user_id    BIGINT NOT NULL REFERENCES users(id)         ON DELETE CASCADE,
  role       TEXT   NOT NULL CHECK (role IN ('owner','member')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (org_id, user_id)
);

-- index'ler
CREATE INDEX idx_org_members_user ON org_members(user_id);
CREATE INDEX idx_org_members_org  ON org_members(org_id);

-- +goose Down
DROP TABLE IF EXISTS org_members;
DROP TABLE IF EXISTS organizations;
