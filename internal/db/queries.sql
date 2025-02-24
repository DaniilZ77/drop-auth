-- name: SaveUser :one
insert into "users" ("username", "pseudonym", "first_name", "last_name")
values ($1, $2, $3, $4)
returning "id";

-- name: UpdateUser :one
update "users"
set "pseudonym" = coalesce(sqlc.narg('pseudonym'), "pseudonym"),
"first_name" = coalesce(sqlc.narg('first_name'), "first_name"),
"last_name" = coalesce(sqlc.narg('last_name'), "last_name"),
"updated_at" = now()
where id = sqlc.arg('id')
and "is_deleted" = false
returning *;

-- name: GetUserByID :one
select * from "users"
where id = $1
and "is_deleted" = false;

-- name: GetUserAdminByUsername :one
select u.id, ua.scale from "users" u
left join "users_admins" ua on u.id = ua.user_id
where u.username = $1
and "is_deleted" = false;

-- name: GetUserAdminByID :one
select u.id, ua.scale from "users" u
left join "users_admins" ua on u.id = ua.user_id
where u.id = $1
and "is_deleted" = false;


-- name: SaveAdmin :exec
insert into "users_admins" ("user_id", "scale") values ($1, $2);

-- name: DeleteAdmin :exec
delete from "users_admins" where user_id = $1;

-- name: GetAdmins :many
select u.id, u.username, ua.scale, ua.created_at
from "users_admins" ua
join "users" u on u.id = ua.user_id
where u.id = coalesce(sqlc.narg('user_id'), u.id)
and u.username = coalesce(sqlc.narg('username'), u.username)
and ua.scale = coalesce(sqlc.narg('admin_scale'), ua.scale)
and u.is_deleted = false
order by ua.created_at
limit sqlc.arg('limit') offset sqlc.arg('offset');

-- name: CountAdmins :one
select count(*)
from "users_admins" ua
join "users" u on u.id = ua.user_id
where u.id = coalesce(sqlc.narg('user_id'), u.id)
and u.username = coalesce(sqlc.narg('username'), u.username)
and ua.scale = coalesce(sqlc.narg('admin_scale'), ua.scale)
and u.is_deleted = false;