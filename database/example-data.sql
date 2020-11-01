-- Bootstrap with some example data
-- Using the reference React app (https://github.com/BryanMorgan/time-tracking-app) you can login with the test user:
-- Email: test@example.com
-- Password: 12345678
insert into account (account_id, company) values (1, 'ACME');
insert into profile (profile_id, email, password, first_name, last_name, timezone) values(1, 'test@example.com', '$2a$12$vH/WnttP4WA7j26tgX4sXuPGMscG9q5ruef1icKsNjoesTMgUEoO2', 'Time', 'Traveler', 'America/Los_Angeles');
insert into profile_account (profile_id, account_id, role) SELECT profile_id, 1, 'admin' from profile where email = 'test@example.com';
insert into client (client_id, account_id, client_name, address) values(1, 1, 'Apple Inc', '1 Infinite Loop Cupertino, CA');
insert into project (project_id, account_id, client_id, project_name, code) VALUES (1, 1, 1, 'iPhone Launch', null);
insert into project (project_id, account_id, client_id, project_name, code) VALUES (2, 1, 2, 'App Development', null);
insert into task (task_id, account_id, task_name, default_rate, default_billable) values (1, 1, 'Development', 100.0, true);
insert into task (task_id, account_id, task_name, default_rate, default_billable) values (2, 1, 'Design', 90.0, false);
insert into project_task(project_id, task_id, account_id, rate, billable) values (1, 1, 1, 95.0, true);
insert into project_task(project_id, task_id, account_id, rate, billable) values (1, 2, 1, 85.0, false);
insert into time (account_id, profile_id, project_id, task_id, day, hours, notes) VALUES (1, 1, 1, 1, '2020-11-01', 4, null);
insert into time (account_id, profile_id, project_id, task_id, day, hours, notes) VALUES  (1, 1, 2, 1, '2020-11-02', 5, null);

ALTER SEQUENCE account_account_id_seq RESTART WITH 2;
ALTER SEQUENCE profile_profile_id_seq RESTART WITH 2;
ALTER SEQUENCE client_client_id_seq RESTART WITH 2;
ALTER SEQUENCE project_project_id_seq RESTART WITH 3;
ALTER SEQUENCE task_task_id_seq RESTART WITH 3;
