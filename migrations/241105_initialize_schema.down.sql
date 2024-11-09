DROP TABLE users;

DROP INDEX idx_users_unique_email;

DROP TABLE wss;

DROP TABLE ws_members;

DROP INDEX idx_ws_members_unique_ws_id_user_id;

DROP INDEX idx_ws_members_ws_id;

DROP TABLE ws_roles;

DROP INDEX idx_ws_roles_unique_ws_id_name;

DROP INDEX idx_ws_roles_ws_id;

DROP TABLE ws_role_members;

DROP INDEX idx_ws_role_members_unique_ws_id_ws_role_id_ws_member_id;

DROP INDEX idx_ws_role_members_ws_id_ws_member_id;

DROP INDEX idx_ws_role_members_ws_id_ws_role_id;

DROP INDEX idx_ws_role_members_ws_id;