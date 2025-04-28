-- CREATE DATABASE goforces;

CREATE TABLE IF NOT EXISTS submissions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    problem_id INT NOT NULL,
    code TEXT NOT NULL,
    tests_status JSONB NOT NULL,
    submission_status VARCHAR(55) NOT NULL
);

CREATE TABLE IF NOT EXISTS problems (
    problem_id SERIAL PRIMARY KEY,
    owner_id INT NOT NULL,
    title VARCHAR(255) NOT NULL,
    statement TEXT NOT NULL,
    time_limit INT NOT NULL,      -- in seconds
    memory_limit INT NOT NULL,    -- in MB
    inputs JSONB NOT NULL,
    outputs JSONB NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('Draft', 'Published', 'Rejected')),
    feedback TEXT,
    publish_date TIMESTAMP
);


CREATE TABLE IF NOT EXISTS users (
    user_id SERIAL PRIMARY KEY,  -- Auto-incrementing ID
    username VARCHAR(100) NOT NULL,  -- Username can't be null
    email VARCHAR(100) NOT NULL UNIQUE,  -- Email is unique and can't be null
    password VARCHAR(255) NOT NULL,  -- Password field
    role VARCHAR(10) CHECK (role IN ('admin', 'user')) NOT NULL  -- Role must be 'admin' or 'user'
);

-- INSERT INTO users (username, email, password, role)
-- VALUES (
--     'adminuser',
--     'admin@example.com',
--     'hashedpassword123',  -- In real apps, use a real hash!
--     'admin'
-- )
-- ON CONFLICT (email) DO NOTHING;

-- INSERT INTO submissions (user_id, problem_id, code, status)
-- VALUES (
--     1,  -- assuming user with ID 1 exists
--     101,  -- example problem ID
--     'print("Hello, world!")',  -- sample code
--     'Accepted'  -- submission status
-- )
-- ON CONFLICT DO NOTHING;
