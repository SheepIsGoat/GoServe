CREATE DATABASE server_db;

\c server_db

CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(32) UNIQUE NOT NULL,
    password VARCHAR(100)
);

CREATE INDEX idx_users_email ON users(email);