-- Time Tracking PostgreSQL Schema

CREATE TABLE IF NOT EXISTS profile
(
    profile_id                 SERIAL PRIMARY KEY,
    email                      TEXT UNIQUE NOT NULL,
    password                   TEXT        NOT NULL,
    first_name                 TEXT        NOT NULL,
    last_name                  TEXT        NOT NULL,
    phone                      TEXT        NULL,
    created                    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated                    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    profile_status             TEXT        NOT NULL DEFAULT 'new',
    timezone                   TEXT        NOT NULL DEFAULT 'America/New_York',

    -- Incorrect password attempts lock
    locked_until               TIMESTAMPTZ NULL,

    -- Forgot password
    forgot_password_token      TEXT        NULL,
    forgot_password_expiration TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS profile_account
(
    profile_id             INT         NOT NULL,
    account_id             INT         NOT NULL,
    profile_account_status TEXT        NOT NULL DEFAULT 'valid',
    role                   TEXT        NOT NULL DEFAULT 'none',
    last_used              TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (profile_id, account_id)
);

CREATE TABLE IF NOT EXISTS account
(
    account_id       SERIAL PRIMARY KEY,
    company          TEXT        NOT NULL,

    account_status   TEXT        NOT NULL DEFAULT 'new',
    week_start       SMALLINT    NOT NULL DEFAULT 1, -- Sunday=0, Monday=1
    account_timezone TEXT        NOT NULL DEFAULT 'America/New_York',
    close_reason     TEXT        NOT NULL DEFAULT '',

    created          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS session
(
    token            TEXT PRIMARY KEY,
    token_expiration TIMESTAMPTZ NOT NULL,
    profile_id       INT         NOT NULL,
    account_id       INT         NOT NULL,
    created          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    type             TEXT        NOT NULL DEFAULT 'web'
);

CREATE INDEX session_profile_idx ON session (profile_id);
CREATE INDEX token_expiration_idx ON session (token_expiration);
CREATE INDEX session_type_idx ON session (type);

CREATE TABLE IF NOT EXISTS login_attempts
(
    attempt_id         SERIAL PRIMARY KEY,
    email              TEXT        NOT NULL,
    login_attempt_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ip_address         INET        NOT NULL
);

CREATE INDEX login_attempts_idx ON login_attempts (email, login_attempt_time);
CREATE INDEX ip_address_idx ON login_attempts (ip_address);

-- Table used by ping requests to validate database connection. Should only contain 1 row with an id of 1
CREATE TABLE ping (ping_id SMALLSERIAL PRIMARY KEY);

CREATE TABLE IF NOT EXISTS client
(
    client_id     SERIAL PRIMARY KEY,
    account_id    INT     NOT NULL,
    client_name   TEXT    NOT NULL,
    address       TEXT    NULL,
    client_active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX client_account_idx ON client (account_id);


CREATE TABLE IF NOT EXISTS project
(
    project_id     SERIAL PRIMARY KEY,
    account_id     INT     NOT NULL,
    client_id      INT     NOT NULL,
    project_name   TEXT    NOT NULL,
    code           TEXT    NULL,
    project_active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX project_account_idx ON project (account_id);


CREATE TABLE IF NOT EXISTS task
(
    task_id          SERIAL PRIMARY KEY,
    account_id       INT            NOT NULL,
    task_name        TEXT           NOT NULL,
    default_rate     NUMERIC(12, 2) NULL,
    default_billable BOOLEAN        NOT NULL DEFAULT TRUE,
    common           BOOLEAN        NOT NULL DEFAULT FALSE,
    task_active      BOOLEAN        NOT NULL DEFAULT TRUE
);

CREATE INDEX task_account_idx ON task (account_id);


CREATE TABLE IF NOT EXISTS project_task
(
    project_id     INT            NOT NULL,
    task_id        INT            NOT NULL,
    account_id     INT            NOT NULL,
    rate           NUMERIC(12, 2) NULL,
    billable       BOOLEAN        NOT NULL DEFAULT TRUE,
    project_active BOOLEAN        NOT NULL DEFAULT TRUE,
    PRIMARY KEY (project_id, task_id)
);

CREATE INDEX project_task_account_idx ON project_task (account_id);


CREATE TABLE IF NOT EXISTS time
(
    account_id INT            NOT NULL,
    profile_id INT            NOT NULL,
    project_id INT            NOT NULL,
    task_id    INT            NOT NULL,
    day        DATE           NOT NULL,
    hours      NUMERIC(12, 2) NOT NULL,
    notes      TEXT           NULL,
    updated    TIMESTAMPTZ    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (account_id, project_id, task_id, profile_id, day)
);

CREATE INDEX time_profile_idx ON time (account_id, profile_id, day DESC);
CREATE INDEX time_task_idx ON time (account_id, task_id);


