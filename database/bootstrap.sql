-- PostgreSQL bootstrap DDL to create a database and a development user account with full access
CREATE DATABASE timetracker
WITH
ENCODING = 'UTF8'
LC_COLLATE = 'en_US.UTF-8'
LC_CTYPE = 'en_US.UTF-8'
TABLESPACE = pg_default
CONNECTION LIMIT = -1;

CREATE ROLE timemaster WITH LOGIN PASSWORD 'timemaster-change-me';

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO timemaster;

