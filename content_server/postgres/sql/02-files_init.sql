\c server_db

CREATE TABLE files (
    unique_filename VARCHAR(255) PRIMARY KEY,
    account_uuid UUID NOT NULL,
    upload_time DATE NOT NULL,
    file_ext VARCHAR(10),
    raw_text TEXT,
    bucket_dir VARCHAR(255),
    location VARCHAR(100)
);

CREATE INDEX idx_files_account_uuid ON files(account_uuid);