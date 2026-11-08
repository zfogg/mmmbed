-- name: Lookup
-- access: read
-- rows: zero-or-one
select id,name from users where id=:id and active=1;
