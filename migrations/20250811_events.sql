-- +goose Up
CREATE TABLE IF NOT EXISTS events (
  id              BIGSERIAL PRIMARY KEY,
  user_id         BIGINT,
  application_id  BIGINT,
  type            TEXT NOT NULL,
  payload_json    JSONB NOT NULL,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_events_user_id ON events(user_id);
CREATE INDEX IF NOT EXISTS idx_events_app_id  ON events(application_id);

-- +goose Down
DROP TABLE IF EXISTS events;
