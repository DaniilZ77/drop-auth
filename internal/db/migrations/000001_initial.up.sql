create extension if not exists "uuid-ossp";

create table if not exists "users" (
    "id" uuid primary key default uuid_generate_v4(),
    "external_id" integer not null,
    "username" varchar(64) not null,
    "pseudonym" varchar(64) not null,
    "first_name" varchar(128) not null,
    "last_name" varchar(128) not null,
    "is_deleted" boolean not null default false,
    "created_at" timestamp not null default now(),
    "updated_at" timestamp not null default now()
);

alter table "users" add constraint "users_username_key" unique (username);
alter table "users" add constraint "users_external_id_key" unique (external_id);
create index on "users" ("pseudonym");
create index on "users" ("username");
create index on "users" ("first_name");
create index on "users" ("last_name");
create index on "users" ("external_id");
