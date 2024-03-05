\c server_db

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE files (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_uuid UUID NOT NULL,
    filepath VARCHAR(255) NOT NULL,
    filename VARCHAR(255),
    upload_time TIMESTAMP NOT NULL,
    file_ext VARCHAR(10),
    raw_text TEXT,
    bucket_dir VARCHAR(255),
    location VARCHAR(100)
);

CREATE INDEX idx_files_account_uuid ON files(account_uuid);