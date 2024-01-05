
-- +migrate Up
CREATE TABLE IF NOT EXISTS sensor (
    id int auto_increment primary key,
    sensor_value float,
    sensor_type varchar(100),
    ID1 varchar(100),
    ID2 int,
    timestamp timestamp default now()
    );
-- +migrate Down
