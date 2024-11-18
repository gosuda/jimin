CREATE EXTENSION vector;

CREATE TABLE users (
    id BIGINT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    email_verified BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_users_unique_email ON users (email);

-- workspaces
CREATE TABLE wss (
    id BIGINT PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ws_members (
    id BIGINT PRIMARY KEY,
    ws_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_ws_members_unique_ws_id_user_id ON ws_members (ws_id, user_id);

CREATE INDEX idx_ws_members_ws_id ON ws_members (ws_id);

CREATE TABLE ws_roles (
    id BIGINT PRIMARY KEY,
    ws_id BIGINT NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_ws_roles_unique_ws_id_name ON ws_roles (ws_id, name);

CREATE INDEX idx_ws_roles_ws_id ON ws_roles (ws_id);

CREATE TABLE ws_role_members (
    id BIGINT PRIMARY KEY,
    ws_id BIGINT NOT NULL,
    ws_role_id BIGINT NOT NULL,
    ws_member_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_ws_role_members_unique_ws_id_ws_role_id_ws_member_id ON ws_role_members (
    ws_id,
    ws_role_id,
    ws_member_id
);

CREATE INDEX idx_ws_role_members_ws_id_ws_member_id ON ws_role_members (ws_id, ws_member_id);

CREATE INDEX idx_ws_role_members_ws_id_ws_role_id ON ws_role_members (ws_id, ws_role_id);

CREATE INDEX idx_ws_role_members_ws_id ON ws_role_members (ws_id);

CREATE TABLE randflake_nodes (
    id BIGSERIAL PRIMARY KEY,
    range_start BIGINT NOT NULL,
    range_end BIGINT NOT NULL,
    valid_from BIGINT NOT NULL,
    valid_to BIGINT NOT NULL,
    lease_holder TEXT NOT NULL
);