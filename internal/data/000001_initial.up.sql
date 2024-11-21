CREATE TABLE IF NOT EXISTS "users" (
    "id" SERIAL PRIMARY KEY,
    "username" VARCHAR(64) NOT NULL,
    "email" VARCHAR(255),
    "telephone" TEXT,
    "pseudonym" VARCHAR(64) NOT NULL,
    "first_name" VARCHAR(128) NOT NULL,
    "last_name" VARCHAR(128) NOT NULL,
    "middle_name" VARCHAR(128),
    "password_hash" VARCHAR(255) NOT NULL,
    "is_deleted" BOOLEAN NOT NULL DEFAULT false,
    "created_at" TIMESTAMP NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT NOW()
);

ALTER TABLE "users" ADD CONSTRAINT "users_username_key" UNIQUE (username);
ALTER TABLE "users" ADD CONSTRAINT "users_email_key" UNIQUE (email);
ALTER TABLE "users" ADD CONSTRAINT "users_telephone_key" UNIQUE (telephone);
CREATE INDEX ON "users" ("pseudonym");
CREATE INDEX ON "users" ("username");
CREATE INDEX ON "users" ("email");
CREATE INDEX ON "users" ("telephone");
