
-- +migrate Up
CREATE TABLE IF NOT EXISTS user (
    id int auto_increment primary key,
    email varchar(100),
    password varchar(100),
    roles varchar(100)
    );
-- +migrate Down
