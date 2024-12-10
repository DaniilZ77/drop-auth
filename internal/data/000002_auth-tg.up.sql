create type "auth_providers" as enum ('tg');

create table "external_users" (
    id serial primary key,
    external_id integer not null,
    user_id integer references "users" ("id"),
    auth_provider auth_providers not null
);

create index on "external_users" ("external_id");

create index on "external_users" ("user_id");
