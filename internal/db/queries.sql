-- name: SaveUser :one
insert into "users" ("username", "external_id", "pseudonym", "first_name", "last_name")
values ($1, $2, $3, $4, $5)
returning "id";

-- name: UpdateUser :one
update "users"
set "pseudonym" = coalesce(sqlc.narg('pseudonym'), "pseudonym"),
"first_name" = coalesce(sqlc.narg('first_name'), "first_name"),
"last_name" = coalesce(sqlc.narg('last_name'), "last_name")
where id = sqlc.arg('id')
and "is_deleted" = false
returning *;

-- name: GetUserByExternalID :one
select * from "users"
where external_id = $1
and "is_deleted" = false;

-- name: GetUserByID :one
select * from "users"
where id = $1
and "is_deleted" = false;

-- name: GetUserByUsername :one
select * from "users"
where username = $1
and "is_deleted" = false;

-- name: SaveAdmin :exec
insert into "users_admins" ("user_id", "scale") values ($1, $2);

-- name: DeleteAdmin :exec
delete from "users_admins" where user_id = $1;

-- name: GetAdminByID :one
select u.id, ua.scale
from "users" u
left join "users_admins" ua on u.id = ua.user_id
where u.id = $1
and "is_deleted" = false;