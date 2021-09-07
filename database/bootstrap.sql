-- PostgreSQL bootstrap DDL to create a database and a development user account with full access
CREATE DATABASE timetracker;
CREATE USER timetraveler WITH PASSWORD  'timetraveler_changeme';

-- Then run: "psql timetracker < schema1.sql"

-- Then run: "psql timetracker" with the following:
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO timetraveler;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO timetraveler;
