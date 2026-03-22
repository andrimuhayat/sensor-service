-- +migrate Up
ALTER TABLE user ADD COLUMN status BOOLEAN DEFAULT true;

-- +migrate Down
ALTER TABLE user DROP COLUMN status;