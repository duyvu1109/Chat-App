CREATE DATABASE IF NOT EXISTS chatapp;
USE chatapp;
CREATE TABLE users (
    name varchar(255),
    room varchar(255),
    msg varchar(255)
);