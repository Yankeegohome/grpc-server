insert into apps (id, name, secret)
values (1, 'tests', 'tests-secret')
on conflict do nothing;