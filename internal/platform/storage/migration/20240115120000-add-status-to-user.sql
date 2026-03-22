-- +migrate Up
ALTER TABLE user ADD COLUMN status varchar(20) DEFAULT 'activated';
-- +migrate Down
ALTER TABLE user DROP COLUMN status;