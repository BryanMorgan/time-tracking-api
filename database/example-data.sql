-- Bootstrap with some example data
insert into account (account_id, company) values (1, 'ACME');
insert into profile (profile_id, email, password, first_name, last_name, timezone) values(1, 'test@example.com', '$2a$12$txCe8h8QH5bYwJC9k5JEr.ZxgHWPo6bxiMHpd8SQ/ETe8mk5.h.jG', 'Time', 'Traveler', 'America/Los_Angeles');
insert into profile_account (profile_id, account_id, role) SELECT profile_id, 1, 'admin' from profile where email = 'test@example.com';
insert into client (client_id, account_id, client_name, address) values(1, 1, 'Apple Inc', '1 Infinite Loop Cupertino, CA');
insert into project (project_id, account_id, client_id, project_name, code) VALUES (1, 1, 1, 'iPhone Launch', null);
insert into project (project_id, account_id, client_id, project_name, code) VALUES (2, 1, 2, 'App Development', null);
insert into task (task_id, account_id, task_name, default_rate, default_billable) values (1, 1, 'Development', 100.0, true);
insert into task (task_id, account_id, task_name, default_rate, default_billable) values (2, 1, 'Design', 90.0, false);
insert into project_task(project_id, task_id, account_id, rate, billable) values (1, 1, 1, 95.0, true);
insert into project_task(project_id, task_id, account_id, rate, billable) values (1, 2, 1, 85.0, false);
insert into time (account_id, profile_id, project_id, task_id, day, hours, notes) VALUES (1, 1, 1, 1, '2020-05-08', 4, null);
insert into time (account_id, profile_id, project_id, task_id, day, hours, notes) VALUES  (1, 1, 2, 1, '2020-05-09', 5, null);

