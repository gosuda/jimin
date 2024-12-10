CREATE EXTENSION vector;

CREATE TABLE
    users (
        id BIGINT PRIMARY KEY,
        name TEXT NOT NULL,
        email TEXT NOT NULL,
        email_verified BOOLEAN NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE UNIQUE INDEX idx_users_unique_email ON users (email);

CREATE TABLE
    users_auth (
        id BIGINT PRIMARY KEY,
        user_id BIGINT NOT NULL,
        provider_id BIGINT NOT NULL,
        provider_subject TEXT NOT NULL,
        associated_data TEXT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE UNIQUE INDEX idx_users_auth_unique_user_id_provider ON users_auth (user_id, provider);
CREATE UNIQUE INDEX idx_users_auth_unique_provider_subject ON users_auth (provider_id, provider_subject);

-- workspaces
CREATE TABLE
    wss (
        id BIGINT PRIMARY KEY,
        name TEXT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE TABLE
    ws_members (
        id BIGINT PRIMARY KEY,
        ws_id BIGINT NOT NULL,
        user_id BIGINT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE UNIQUE INDEX idx_ws_members_unique_ws_id_user_id ON ws_members (ws_id, user_id);

CREATE INDEX idx_ws_members_ws_id ON ws_members (ws_id);

CREATE TABLE
    ws_roles (
        id BIGINT PRIMARY KEY,
        ws_id BIGINT NOT NULL,
        name TEXT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE UNIQUE INDEX idx_ws_roles_unique_ws_id_name ON ws_roles (ws_id, name);

CREATE INDEX idx_ws_roles_ws_id ON ws_roles (ws_id);

CREATE TABLE
    ws_role_members (
        id BIGINT PRIMARY KEY,
        ws_id BIGINT NOT NULL,
        ws_role_id BIGINT NOT NULL,
        ws_member_id BIGINT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE UNIQUE INDEX idx_ws_role_members_unique_ws_id_ws_role_id_ws_member_id ON ws_role_members (ws_id, ws_role_id, ws_member_id);

CREATE INDEX idx_ws_role_members_ws_id_ws_member_id ON ws_role_members (ws_id, ws_member_id);

CREATE INDEX idx_ws_role_members_ws_id_ws_role_id ON ws_role_members (ws_id, ws_role_id);

CREATE INDEX idx_ws_role_members_ws_id ON ws_role_members (ws_id);

CREATE TYPE relation_type AS ENUM ('INCLUDE', 'REWRITE', 'OTHER');

CREATE TABLE
    ws_relations (
        id BIGINT PRIMARY KEY,
        ws_id BIGINT NOT NULL,
        object_id BIGINT NOT NULL,
        relation relation_type NOT NULL,
        target_id BIGINT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
    );

CREATE TABLE
    randflake_nodes (
        id BIGSERIAL PRIMARY KEY,
        range_start BIGINT NOT NULL,
        range_end BIGINT NOT NULL,
        lease_holder UUID NOT NULL,
        lease_start BIGINT NOT NULL,
        lease_end BIGINT NOT NULL
    );

CREATE UNIQUE INDEX idx_randflake_nodes_unique_range_start ON randflake_nodes (range_start);

CREATE UNIQUE INDEX idx_randflake_nodes_unique_range_start_lease_end ON randflake_nodes (range_start ASC, lease_end DESC);

CREATE INDEX idx_randflake_nodes_lease_end ON randflake_nodes (lease_end ASC);