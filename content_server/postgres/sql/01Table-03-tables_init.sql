\c server_db

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE "SampleAccounts" (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    avatar VARCHAR(255),
    name VARCHAR(32),
    title VARCHAR (64)
);

CREATE TABLE "SampleInvoices" (
    id SERIAL PRIMARY KEY,
    account_id UUID,
    amount FLOAT,
    status VARCHAR(32),
    date DATE
)