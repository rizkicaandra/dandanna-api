-- Migration 001: Create user schema
-- Creates the full user schema with all tables, enums, and triggers.

CREATE SCHEMA IF NOT EXISTS "user";

-- Required on PostgreSQL < 13 for gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ─── Enums ───────────────────────────────────────────────────────────────────

CREATE TYPE "user".user_role_status AS ENUM ('active', 'inactive', 'suspended');
CREATE TYPE "user".primary_service   AS ENUM ('makeup', 'hair', 'attire');

-- ─── updated_at trigger ───────────────────────────────────────────────────────

CREATE OR REPLACE FUNCTION "user".set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ─── Tables ───────────────────────────────────────────────────────────────────

CREATE TABLE "user".application (
    id         bigserial    PRIMARY KEY,
    code       char(20)     NOT NULL UNIQUE,
    name       varchar(100) NOT NULL UNIQUE,
    created_at timestamptz  NOT NULL DEFAULT NOW(),
    updated_at timestamptz  NOT NULL DEFAULT NOW(),
    deleted_at timestamptz,
    created_by varchar(255),
    updated_by varchar(255),
    deleted_by varchar(255)
);

CREATE TABLE "user".module (
    id          bigserial    PRIMARY KEY,
    code        char(20)     NOT NULL UNIQUE,
    name        varchar(100) NOT NULL UNIQUE,
    description varchar(255),
    created_at  timestamptz  NOT NULL DEFAULT NOW(),
    updated_at  timestamptz  NOT NULL DEFAULT NOW(),
    deleted_at  timestamptz,
    created_by  varchar(255),
    updated_by  varchar(255),
    deleted_by  varchar(255)
);

CREATE TABLE "user".submodule (
    id          bigserial    PRIMARY KEY,
    code        char(20)     NOT NULL,
    name        varchar(100) NOT NULL,
    description varchar(255),
    module_id   bigint       NOT NULL REFERENCES "user".module(id),
    created_at  timestamptz  NOT NULL DEFAULT NOW(),
    updated_at  timestamptz  NOT NULL DEFAULT NOW(),
    deleted_at  timestamptz,
    created_by  varchar(255),
    updated_by  varchar(255),
    deleted_by  varchar(255),
    UNIQUE (module_id, code),
    UNIQUE (module_id, name)
);

CREATE TABLE "user".role (
    id             bigserial    PRIMARY KEY,
    code           varchar(50)  NOT NULL UNIQUE,
    name           varchar(100) NOT NULL,
    application_id bigint       NOT NULL REFERENCES "user".application(id),
    created_at     timestamptz  NOT NULL DEFAULT NOW(),
    updated_at     timestamptz  NOT NULL DEFAULT NOW(),
    deleted_at     timestamptz,
    created_by     varchar(255),
    updated_by     varchar(255),
    deleted_by     varchar(255)
);

CREATE TABLE "user".permission (
    id           bigserial   PRIMARY KEY,
    role_id      bigint      NOT NULL REFERENCES "user".role(id),
    module_id    bigint      NOT NULL REFERENCES "user".module(id),
    submodule_id bigint      NOT NULL REFERENCES "user".submodule(id),
    action       jsonb       NOT NULL DEFAULT '{}',
    created_at   timestamptz NOT NULL DEFAULT NOW(),
    updated_at   timestamptz NOT NULL DEFAULT NOW(),
    deleted_at   timestamptz,
    created_by   varchar(255),
    updated_by   varchar(255),
    deleted_by   varchar(255),
    UNIQUE (role_id, module_id, submodule_id)
);

CREATE TABLE "user".user_management (
    id         uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    name       varchar(255) NOT NULL,
    phone      varchar(20)  NOT NULL,
    email      varchar(255) NOT NULL UNIQUE,
    password   text         NOT NULL,
    created_at timestamptz  NOT NULL DEFAULT NOW(),
    updated_at timestamptz  NOT NULL DEFAULT NOW(),
    deleted_at timestamptz,
    created_by varchar(255),
    updated_by varchar(255),
    deleted_by varchar(255)
);

CREATE TABLE "user".user_role (
    id                 bigserial                  PRIMARY KEY,
    user_management_id uuid                       NOT NULL REFERENCES "user".user_management(id),
    role_id            bigint                     NOT NULL REFERENCES "user".role(id),
    application_id     bigint                     NOT NULL REFERENCES "user".application(id),
    status             "user".user_role_status    NOT NULL DEFAULT 'active',
    created_at         timestamptz                NOT NULL DEFAULT NOW(),
    updated_at         timestamptz                NOT NULL DEFAULT NOW(),
    deleted_at         timestamptz,
    created_by         varchar(255),
    updated_by         varchar(255),
    deleted_by         varchar(255),
    UNIQUE (user_management_id, role_id, application_id)
);

CREATE TABLE "user".artist_profile (
    id                 uuid                    PRIMARY KEY DEFAULT gen_random_uuid(),
    user_management_id uuid                    NOT NULL UNIQUE REFERENCES "user".user_management(id),
    business_name      varchar(255)            NOT NULL,
    primary_service    "user".primary_service  NOT NULL,
    city               varchar(100)            NOT NULL,
    instagram          varchar(100),
    created_at         timestamptz             NOT NULL DEFAULT NOW(),
    updated_at         timestamptz             NOT NULL DEFAULT NOW(),
    deleted_at         timestamptz,
    created_by         varchar(255),
    updated_by         varchar(255),
    deleted_by         varchar(255)
);

-- ─── updated_at triggers ──────────────────────────────────────────────────────

CREATE TRIGGER trg_application_updated_at
    BEFORE UPDATE ON "user".application
    FOR EACH ROW EXECUTE FUNCTION "user".set_updated_at();

CREATE TRIGGER trg_module_updated_at
    BEFORE UPDATE ON "user".module
    FOR EACH ROW EXECUTE FUNCTION "user".set_updated_at();

CREATE TRIGGER trg_submodule_updated_at
    BEFORE UPDATE ON "user".submodule
    FOR EACH ROW EXECUTE FUNCTION "user".set_updated_at();

CREATE TRIGGER trg_role_updated_at
    BEFORE UPDATE ON "user".role
    FOR EACH ROW EXECUTE FUNCTION "user".set_updated_at();

CREATE TRIGGER trg_permission_updated_at
    BEFORE UPDATE ON "user".permission
    FOR EACH ROW EXECUTE FUNCTION "user".set_updated_at();

CREATE TRIGGER trg_user_management_updated_at
    BEFORE UPDATE ON "user".user_management
    FOR EACH ROW EXECUTE FUNCTION "user".set_updated_at();

CREATE TRIGGER trg_user_role_updated_at
    BEFORE UPDATE ON "user".user_role
    FOR EACH ROW EXECUTE FUNCTION "user".set_updated_at();

CREATE TRIGGER trg_artist_profile_updated_at
    BEFORE UPDATE ON "user".artist_profile
    FOR EACH ROW EXECUTE FUNCTION "user".set_updated_at();
