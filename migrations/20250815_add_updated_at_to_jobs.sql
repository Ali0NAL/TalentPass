-- +migrate Up
ALTER TABLE jobs
ADD COLUMN updated_at TIMESTAMPTZ DEFAULT now();

-- +migrate Down
ALTER TABLE jobs
DROP COLUMN updated_at;
