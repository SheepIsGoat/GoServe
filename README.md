# GoServe
# create your postgres instance
`docker run --name my-shades-postgres -e POSTGRES_PASSWORD=mysecretpassword -d -p 5432:5432 postgres`
`docker exec -it my-shades-postgres psql -U postgres`

# Create your database and table
```
CREATE DATABASE server_db;
\c server_db
CREATE TABLE users (
    username VARCHAR(32) PRIMARY KEY,
    password VARCHAR(100)
);
CREATE TABLE content ( 
    content_id SERIAL PRIMARY KEY, 
    title VARCHAR(64), 
    summary VARCHAR(256), 
    background_url VARCHAR(128),
    datetime TIMESTAMP
);
INSERT INTO content (
        content_id,
        title,
        summary,
        background_url,
        datetime
    ) VALUES (
        1,
        'Dogs strike back',
        'Are dogs bad, or just fluffy friends?',
        'https://amazon.s3.us-east-1.120-48fj-Gjf',
        '2023-09-19 01:50:00'
    )
```

# Create your user
`curl -X POST -d "username=gopher&password=G0ph3r" http://localhost:8080/create_user`

# Login as your user
`curl -X POST -d "username=gopher&password=G0ph3r" http://localhost:8080/login`

# Request content
`curl -H "Content-Type: application/json" -H "Authorization: Bearer YOUR_TOKEN_HERE" -d "{\"time\":\"2023-09-18T06:36:00\"}" http://localhost:8080/new_content_p`
