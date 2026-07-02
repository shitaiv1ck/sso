CREATE SCHEMA sso;

CREATE TABLE sso.users(
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE, 
    pass_hash VARCHAR(255) NOT NULL,

    CHECK(email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);
CREATE INDEX idx_users_email ON sso.users(email);

CREATE TABLE sso.apps(
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE CHECK (char_length(name) BETWEEN 2 AND 255)
);
CREATE INDEX idx_apps_name ON sso.apps(name);